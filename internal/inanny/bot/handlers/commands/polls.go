package command

import (
	"fmt"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/features/polls"
)

type PollsCommandHandler struct {
	CommandHandler
}

func (handler PollsCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	user, err := polls.UpsertUser(polls.UserDto{
		TelegramLogin: update.Message.From.UserName,
		FirstName:     update.Message.From.FirstName,
		LastName:      update.Message.From.LastName,
	})
	if err != nil {
		return err
	}

	args := strings.Fields(update.Message.CommandArguments())
	if len(args) > 0 && args[0] == "delete" {
		return handleDeletePoll(bot, update, user.ID, args)
	}

	return handleListPolls(bot, update, user.ID)
}

func handleListPolls(bot *tgbot.BotAPI, update *tgbot.Update, userID int64) error {
	storedPolls, err := polls.GetCronPollsByChatAndUser(update.Message.Chat.ID, userID)
	if err != nil {
		return err
	}

	if len(storedPolls) == 0 {
		return sendReply(bot, update, "No cron polls in this chat")
	}

	lines := []string{"Cron polls:"}
	for _, storedPoll := range storedPolls {
		lines = append(lines, fmt.Sprintf(
			"#%d [%s] {%s} %s",
			storedPoll.ID,
			storedPoll.Command,
			storedPoll.CronExpr,
			storedPoll.Title,
		))
	}

	return sendReply(bot, update, strings.Join(lines, "\n"))
}

func handleDeletePoll(bot *tgbot.BotAPI, update *tgbot.Update, userID int64, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Usage: /polls delete <id>")
	}

	pollID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid poll id: %s", args[1])
	}

	deleted, err := polls.DeleteCronPollByIDChatAndUser(pollID, update.Message.Chat.ID, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return fmt.Errorf("Cron poll #%d doesn't exist in this chat for current user", pollID)
	}

	polls.RemoveCronPoll(pollID)
	return sendReply(bot, update, fmt.Sprintf("Cron poll #%d was deleted", pollID))
}

func sendReply(bot *tgbot.BotAPI, update *tgbot.Update, text string) error {
	messageConfig := tgbot.NewMessage(update.Message.Chat.ID, text)
	messageConfig.ReplyToMessageID = update.Message.MessageID
	_, err := bot.Send(messageConfig)
	return err
}

func NewPollsCommandHandler() PollsCommandHandler {
	return PollsCommandHandler{
		CommandHandler: CommandHandler{
			command: "polls",
		},
	}
}
