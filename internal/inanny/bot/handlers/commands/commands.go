package command

import (
	"errors"
	"fmt"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	handlers "github.com/holeyko/innany-tgbot/internal/inanny/bot/handlers"
	customcommands "github.com/holeyko/innany-tgbot/internal/inanny/features/customcommands"
)

const maxCustomCommandDepth = 10

type CommandHandler struct {
	command string
}

func (handler CommandHandler) IsSutable(command *string) bool {
	return *command == handler.command
}

var commandHandlers = [...]handlers.TgUpdateHandler[string]{
	NewHelpCommandHandler(),
	NewPollCommandHandler(),
	NewBPCommandHandler(),
	NewTPCommandHandler(),
	NewPollsCommandHandler(),
	NewCommandsCommandHandler(),
	NewHelloCommandHandler(),
}

func FindCommandHandler(command string) handlers.TgUpdateHandler[string] {
	for _, handler := range commandHandlers {
		if handler.IsSutable(&command) {
			return handler
		}
	}

	return nil
}

func IsBuiltInCommand(command string) bool {
	return FindCommandHandler(strings.TrimPrefix(command, "/")) != nil
}

func IsReservedCommand(command string) bool {
	return IsBuiltInCommand(command)
}

func HandleCustomCommand(bot *tgbot.BotAPI, update *tgbot.Update, command string) (bool, error) {
	return handleCustomCommand(bot, update, command, strings.Fields(update.Message.CommandArguments()), 0)
}

func handleCustomCommand(bot *tgbot.BotAPI, update *tgbot.Update, command string, args []string, depth int) (bool, error) {
	if depth >= maxCustomCommandDepth {
		return true, errors.New("Custom command chain is too deep")
	}

	storedCommand, err := customcommands.GetCustomCommandByChatAndName(update.Message.Chat.ID, command)
	if err != nil {
		if errors.Is(err, customcommands.ErrCustomCommandNotFound) {
			return false, nil
		}
		return true, err
	}

	body, err := customcommands.ApplyParameters(storedCommand, args)
	if err != nil {
		return true, err
	}

	if handler := FindCommandHandler(storedCommand.TargetCommand); handler != nil {
		return true, handler.Handle(bot, buildCommandUpdate(update, storedCommand.TargetCommand, body))
	}

	targetArgs := strings.Fields(body)
	handled, err := handleCustomCommand(bot, update, storedCommand.TargetCommand, targetArgs, depth+1)
	if !handled {
		return true, fmt.Errorf("Target command /%s doesn't exist", storedCommand.TargetCommand)
	}
	return true, err
}

func buildCommandUpdate(update *tgbot.Update, command string, body string) *tgbot.Update {
	commandText := "/" + command
	text := commandText
	if body != "" {
		text += " " + body
	}

	message := *update.Message
	message.Text = text
	message.Entities = []tgbot.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(commandText)}}

	commandUpdate := *update
	commandUpdate.Message = &message
	return &commandUpdate
}

func EnsureExistingCommand(chatID int64, command string) error {
	command = strings.TrimPrefix(command, "/")
	if IsBuiltInCommand(command) {
		return nil
	}
	if _, err := customcommands.GetCustomCommandByChatAndName(chatID, command); err != nil {
		if errors.Is(err, customcommands.ErrCustomCommandNotFound) {
			return fmt.Errorf("Existing command /%s doesn't exist", command)
		}
		return err
	}
	return nil
}
