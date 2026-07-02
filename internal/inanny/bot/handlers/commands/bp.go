package command

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bpOptions = []string{"Да", "Нет"}

type BPCommandHandler struct {
	CommandHandler
}

func (handler BPCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	return handlePollCommand(bot, update, "bp", bpOptions)
}

func NewBPCommandHandler() BPCommandHandler {
	return BPCommandHandler{
		CommandHandler: CommandHandler{command: "bp"},
	}
}
