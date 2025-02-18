package command

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HelloCommandHandler struct {
	CommandHandler
}

func (handler HelloCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	firstname := update.Message.From.FirstName
	message := "Привет, мой дорогой друг, " + firstname

	messageConfig := tgbot.NewMessage(update.Message.Chat.ID, message)
	messageConfig.ReplyToMessageID = update.Message.MessageID

	bot.Send(messageConfig)
	return nil
}

func NewHelloCommandHandler() HelloCommandHandler {
	return HelloCommandHandler{
		CommandHandler{
			command: "hello",
		},
	}
}
