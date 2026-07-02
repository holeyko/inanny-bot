package command

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	tpOptions = []string{"Да", "Нет", "Тык"}
)

type TPCommandHandler struct {
	CommandHandler
}

func (handler TPCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	return handlePollCommand(bot, update, "tp", tpOptions)
}

func NewTPCommandHandler() TPCommandHandler {
	return TPCommandHandler{
		CommandHandler: CommandHandler{
			command: "tp",
		},
	}
}
