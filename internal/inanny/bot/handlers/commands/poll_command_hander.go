package command

import (
	"errors"
	"slices"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/bot/polls"
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

func NewPollCommandHandler() PollCommandHandler {
	return PollCommandHandler{
		CommandHandler: CommandHandler{
			command: "poll",
		},
	}
}
