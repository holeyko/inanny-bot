package command

import (
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

	err = polls.SendPoll(bot, &poll, update.Message)
	return err
}

func NewPollCommandHandler() PollCommandHandler {
	return PollCommandHandler{
		CommandHandler: CommandHandler{
			command: "poll",
		},
	}
}
