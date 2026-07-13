package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	customcommands "github.com/holeyko/innany-tgbot/internal/inanny/features/customcommands"
)

type CommandsCommandHandler struct {
	CommandHandler
}

func (handler CommandsCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	args := strings.Fields(update.Message.CommandArguments())
	if len(args) > 0 {
		switch args[0] {
		case "create":
			return handleCreateCustomCommandPrompt(bot, update)
		case "delete":
			return handleDeleteCustomCommand(bot, update, args)
		default:
			return fmt.Errorf("Usage: /commands [create|delete <id>]")
		}
	}

	return handleListCustomCommands(bot, update)
}

func handleCreateCustomCommandPrompt(bot *tgbot.BotAPI, update *tgbot.Update) error {
	user, err := customcommands.UpsertUser(customCommandUserDto(update.Message.From))
	if err != nil {
		return err
	}

	prompt, err := sendReplyMessage(bot, update, "Please provide the command template.")
	if err != nil {
		return err
	}

	_, err = customcommands.CreateCustomCommandDraft(update.Message.Chat.ID, user.ID, int64(prompt.MessageID))
	return err
}

func handleListCustomCommands(bot *tgbot.BotAPI, update *tgbot.Update) error {
	commands, err := customcommands.ListCustomCommandsByChat(update.Message.Chat.ID)
	if err != nil {
		return err
	}

	if len(commands) == 0 {
		return sendReply(bot, update, "No custom commands in this chat")
	}

	lines := []string{"Custom commands:"}
	for _, command := range commands {
		body := strings.ReplaceAll(command.Body, "\n", "\\n")
		lines = append(lines, fmt.Sprintf("#%d /%s -> /%s %s", command.ID, command.Name, command.TargetCommand, body))
	}

	return sendReply(bot, update, strings.Join(lines, "\n"))
}

func handleDeleteCustomCommand(bot *tgbot.BotAPI, update *tgbot.Update, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Usage: /commands delete <id>")
	}

	commandID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid command id: %s", args[1])
	}

	user, err := customcommands.UpsertUser(customCommandUserDto(update.Message.From))
	if err != nil {
		return err
	}

	deleted, err := customcommands.DeleteCustomCommandByIDChatAndUser(commandID, update.Message.Chat.ID, user.ID)
	if err != nil {
		return err
	}
	if deleted {
		return sendReply(bot, update, fmt.Sprintf("Custom command #%d was deleted", commandID))
	}

	admin, err := isChatAdmin(bot, update)
	if err != nil {
		return err
	}
	if !admin {
		return fmt.Errorf("Custom command #%d doesn't exist or you don't have permission to delete it", commandID)
	}

	deleted, err = customcommands.DeleteCustomCommandByIDAndChat(commandID, update.Message.Chat.ID)
	if err != nil {
		return err
	}
	if !deleted {
		return fmt.Errorf("Custom command #%d doesn't exist in this chat", commandID)
	}

	return sendReply(bot, update, fmt.Sprintf("Custom command #%d was deleted", commandID))
}

func TryHandleCustomCommandDraftReply(bot *tgbot.BotAPI, update *tgbot.Update) (bool, error) {
	message := update.Message
	if message == nil || message.From == nil || message.ReplyToMessage == nil {
		return false, nil
	}
	if message.From.UserName == "" {
		return false, nil
	}

	user, err := customcommands.UpsertUser(customCommandUserDto(message.From))
	if err != nil {
		return false, err
	}

	draft, err := customcommands.GetCustomCommandDraftByPromptMessageID(message.Chat.ID, user.ID, int64(message.ReplyToMessage.MessageID))
	if err != nil {
		if errors.Is(err, customcommands.ErrCustomCommandDraftNotFound) {
			return false, nil
		}
		return true, err
	}

	if draft.IsExpired() {
		if err := customcommands.DeleteCustomCommandDraftByID(draft.ID); err != nil {
			return true, err
		}
		return true, errors.New("Custom command draft expired, start again")
	}

	template, err := customcommands.ParseCommandTemplate(message.Text)
	if err != nil {
		return true, err
	}
	if IsReservedCommand(template.Name) {
		return true, fmt.Errorf("Command name /%s is reserved", template.Name)
	}
	if template.Name == template.TargetCommand {
		return true, errors.New("Custom command can't target itself")
	}
	if _, err := customcommands.GetCustomCommandByChatAndName(message.Chat.ID, template.Name); err == nil {
		return true, fmt.Errorf("Command /%s already exists in this chat", template.Name)
	} else if !errors.Is(err, customcommands.ErrCustomCommandNotFound) {
		return true, err
	}
	if err := EnsureExistingCommand(message.Chat.ID, template.TargetCommand); err != nil {
		return true, err
	}

	_, err = customcommands.CreateCustomCommand(customcommands.CreateCustomCommandDto{
		ChatID:        message.Chat.ID,
		UserID:        user.ID,
		Name:          template.Name,
		TargetCommand: template.TargetCommand,
		Body:          template.Body,
	})
	if err != nil {
		return true, err
	}

	if err := customcommands.DeleteCustomCommandDraftByID(draft.ID); err != nil {
		return true, err
	}

	return true, sendReply(bot, update, fmt.Sprintf("Custom command /%s was created", template.Name))
}

func isChatAdmin(bot *tgbot.BotAPI, update *tgbot.Update) (bool, error) {
	member, err := bot.GetChatMember(tgbot.GetChatMemberConfig{ChatConfigWithUser: tgbot.ChatConfigWithUser{
		ChatID: update.Message.Chat.ID,
		UserID: update.Message.From.ID,
	}})
	if err != nil {
		return false, err
	}

	return member.Status == "creator" || member.Status == "administrator", nil
}

func customCommandUserDto(user *tgbot.User) customcommands.UserDto {
	return customcommands.UserDto{
		TelegramLogin: user.UserName,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
	}
}

func sendReplyMessage(bot *tgbot.BotAPI, update *tgbot.Update, text string) (*tgbot.Message, error) {
	messageConfig := tgbot.NewMessage(update.Message.Chat.ID, text)
	messageConfig.ReplyToMessageID = update.Message.MessageID
	message, err := bot.Send(messageConfig)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func NewCommandsCommandHandler() CommandsCommandHandler {
	return CommandsCommandHandler{CommandHandler: CommandHandler{command: "commands"}}
}
