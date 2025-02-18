package command

import (
	"errors"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/polls"
)

type PollCommandHandler struct {
	CommandHandler
}

func (handler PollCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	poll, err := polls.ParsePoll(update.Message.CommandArguments())
	if err != nil {
		return nil
	}
	if len(poll.Options) < 2 {
		return errors.New("Should be at least 2 options")
	}

	pollConfig := tgbot.NewPoll(
		update.Message.Chat.ID,
		poll.Title,
		poll.Options...,
	)

	applyFlagsToPollConfig(&pollConfig, poll.Flags)
	_, err = bot.Send(pollConfig)
	return err
}

func NewPollCommandHandler() PollCommandHandler {
	return PollCommandHandler{
		CommandHandler: CommandHandler{
			command: "poll",
		},
	}
}
