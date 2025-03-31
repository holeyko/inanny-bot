package poll

import (
	"errors"
	"slices"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendPoll(bot *tgbot.BotAPI, poll *Poll, chatId int64) (err error) {
	if len(poll.Options) < 2 {
		return errors.New("Should be at least 2 options")
	}

	pollConfig := tgbot.NewPoll(
		chatId,
		poll.Title,
		poll.Options...,
	)

	applyFlagsToPollConfig(&pollConfig, poll.Flags)
	message, err := bot.Send(pollConfig)

	if err != nil {
		return nil
	}

	if slices.Contains(poll.Flags, Pin) {
		pinConfig := createPinConfig(
			chatId,
			message.MessageID,
			true,
		)

		_, err = bot.Request(pinConfig)
	}

	return
}

func applyFlagsToPollConfig(pollConfig *tgbot.SendPollConfig, flags []Flag) {
	pollConfig.IsAnonymous = false
	for _, flag := range flags {
		switch flag {
		case Anonymous:
			pollConfig.IsAnonymous = true
		case Multipoll:
			pollConfig.AllowsMultipleAnswers = true
		}
	}
}

func createPinConfig(chatId int64, messageId int, notify bool) *tgbot.PinChatMessageConfig {
	return &tgbot.PinChatMessageConfig{
		ChatID:              chatId,
		MessageID:           messageId,
		DisableNotification: !notify,
	}
}
