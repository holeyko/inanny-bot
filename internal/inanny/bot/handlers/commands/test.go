package command

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TestCommandHandler struct {
	CommandHandler
}

func (handler TestCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	firstname := update.Message.From.FirstName
	message := "Привет, мой дорогой друг, " + firstname

	messageConfig := tgbot.NewMessage(update.Message.Chat.ID, message)
	messageConfig.ReplyToMessageID = update.Message.MessageID

	_, err := bot.Send(messageConfig)
	return err
}

func NewTestCommandHandler() TestCommandHandler {
	return TestCommandHandler{
		CommandHandler{
			command: "test",
		},
	}
}
