package command

import (
	"slices"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/bot/polls"
)

const (
	yes   = "Да"
	no    = "Нет"
	other = "Тык"
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
		[]string{yes, no, other}...,
	)

	applyFlagsToPollConfig(&pollConfig, poll.Flags)
	message, err := bot.Send(pollConfig)

	if err != nil {
		return nil
	}

	if slices.Contains(poll.Flags, polls.Pin) {
		pinConfig := createPinConfig(
			update.Message.Chat.ID,
			message.MessageID,
			true,
		)

		_, err = bot.Request(pinConfig)
	}

	return err
}

func NewBinPollCommandHandler() BinPollCommandHandler {
	return BinPollCommandHandler{
		CommandHandler: CommandHandler{
			command: "bin_poll",
		},
	}
}
