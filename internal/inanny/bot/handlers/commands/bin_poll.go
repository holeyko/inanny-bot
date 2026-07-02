package command

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var binPollOptions = []string{"Да", "Нет"}

type BinPollCommandHandler struct {
	CommandHandler
}

func (handler BinPollCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	return handlePollCommand(bot, update, "bin_poll", binPollOptions)
}

func NewBinPollCommandHandler() BinPollCommandHandler {
	return BinPollCommandHandler{
		CommandHandler: CommandHandler{command: "bin_poll"},
	}
}
