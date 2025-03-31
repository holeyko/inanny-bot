package command

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/bot/polls"
)

var (
	binOptions = []string{"Да", "Нет", "Тык"}
)

type BinPollCommandHandler struct {
	CommandHandler
}

func (handler BinPollCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	poll, err := polls.ParsePoll(update.Message.CommandArguments())
	if err != nil {
		return err
	}

	poll.Options = binOptions
	err = polls.SendPoll(bot, &poll, update.Message.Chat.ID)
	return err
}

func NewBinPollCommandHandler() BinPollCommandHandler {
	return BinPollCommandHandler{
		CommandHandler: CommandHandler{
			command: "bin_poll",
		},
	}
}
