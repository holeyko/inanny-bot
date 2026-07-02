package command

import (
	"fmt"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/features/polls"
)

type PollCommandHandler struct {
	CommandHandler
}

func (handler PollCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	return handlePollCommand(bot, update, "poll", nil)
}

func handlePollCommand(bot *tgbot.BotAPI, update *tgbot.Update, command string, options []string) error {
	pollCommand, err := polls.ParsePollCommand(update.Message.CommandArguments())
	if err != nil {
		return err
	}

	if options != nil {
		pollCommand.Poll.Options = options
	}

	if pollCommand.CronExpr == "" {
		return polls.SendPoll(bot, &pollCommand.Poll, update.Message)
	}
	if err := polls.CheckPoll(&pollCommand.Poll); err != nil {
		return err
	}

	user, err := polls.UpsertUser(polls.UserDto{
		TelegramLogin: update.Message.From.UserName,
		FirstName:     update.Message.From.FirstName,
		LastName:      update.Message.From.LastName,
	})
	if err != nil {
		return err
	}

	storedPoll, err := polls.CreateStoredPoll(polls.CreateStoredPollDto{
		ChatID:   update.Message.Chat.ID,
		UserID:   user.ID,
		Command:  command,
		Poll:     pollCommand.Poll,
		CronExpr: pollCommand.CronExpr,
	})
	if err != nil {
		return err
	}

	if err := polls.RegisterCronPoll(storedPoll); err != nil {
		return err
	}

	messageConfig := tgbot.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Cron poll #%d was created", storedPoll.ID))
	messageConfig.ReplyToMessageID = update.Message.MessageID
	_, err = bot.Send(messageConfig)
	return err
}

func NewPollCommandHandler() PollCommandHandler {
	return PollCommandHandler{
		CommandHandler: CommandHandler{
			command: "poll",
		},
	}
}
