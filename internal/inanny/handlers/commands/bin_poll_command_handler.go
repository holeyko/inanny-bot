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

func (handler BinPollCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	poll, err := polls.ParsePoll(update.Message.CommandArguments())
	if err != nil {
		return err
	}

	pollConfig := tgbot.NewPoll(
		update.Message.Chat.ID,
		poll.Title,
		[]string{yes, no}...,
	)

	applyFlagsToPollConfig(&pollConfig, poll.Flags)
	_, err = bot.Send(pollConfig)
	return err
}

func NewBinPollCommandHandler() BinPollCommandHandler {
	return BinPollCommandHandler{
		CommandHandler: CommandHandler{
			command: "bin_poll",
		},
	}
}
