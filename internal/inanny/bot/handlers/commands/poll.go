package command

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/features/polls"
)

type PollCommandHandler struct {
	CommandHandler
}

func (handler PollCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	return handlePollCommand(bot, update, "poll", nil)
}

func handlePollCommand(bot *tgbot.BotAPI, update *tgbot.Update, command string, options []string) error {
	pollCommand, err := polls.ParsePollCommand(update.Message.CommandArguments(), options)
	if err != nil {
		return err
	}

	pollCommand.Poll.Command = command
	return polls.StartPollFlow(bot, update, pollCommand.Poll)
}

func NewPollCommandHandler() PollCommandHandler {
	return PollCommandHandler{
		CommandHandler: CommandHandler{
			command: "poll",
		},
	}
}
