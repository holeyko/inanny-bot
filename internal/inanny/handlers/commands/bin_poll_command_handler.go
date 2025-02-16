package command

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/polls"
)

const (
	yes = "Да"
	no  = "Нет"
)

type BinPollCommandHandler struct {
	CommandHandler
}

func (handler *BinPollCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	poll, _ := polls.ParsePoll(update.Message.CommandArguments())
	pollConfig := tgbot.NewPoll(
		update.Message.Chat.ID,
		poll.Title,
		[]string{yes, no}...,
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

func NewBinPollCommandHandler() *BinPollCommandHandler {
	return &BinPollCommandHandler{
		CommandHandler: CommandHandler{
			command: "bin_poll",
		},
	}
}
