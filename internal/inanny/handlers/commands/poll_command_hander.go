package command

import (
	"errors"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/polls"
)

type PollCommandHandler struct {
	CommandHandler
}

func (handler *PollCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	poll, _ := polls.ParsePoll(update.Message.CommandArguments())
	if len(poll.Options) < 2 {
		return errors.New("Should be at least 2 options")
	}

	pollConfig := tgbot.NewPoll(
		update.Message.Chat.ID,
		poll.Title,
		poll.Options...,
	)

	pollConfig.IsAnonymous = false
	for _, flag := range poll.Flags {
		switch flag {
		case polls.Anonymous:
			pollConfig.IsAnonymous = true
		case polls.Multipoll:
			pollConfig.AllowsMultipleAnswers = true
		}
	}

	bot.Send(pollConfig)
	return nil
}

func NewPollCommandHandler() *PollCommandHandler {
	return &PollCommandHandler{
		CommandHandler: CommandHandler{
			command: "poll",
		},
	}
}
