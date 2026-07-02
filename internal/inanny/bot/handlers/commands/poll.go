package command

import (
	"fmt"

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
	pollCommand, err := polls.ParsePollCommand(update.Message.CommandArguments())
	if err != nil {
		return err
	}

	if options != nil {
		pollCommand.Poll.Options = options
	}

	if pollCommand.CronExpr == "" {
		return polls.SendPoll(bot, &pollCommand.Poll, update.Message)
	}
	return fmt.Errorf("Cron poll creation moved to /bin_poll [cron]")
}

func NewPollCommandHandler() PollCommandHandler {
	return PollCommandHandler{
		CommandHandler: CommandHandler{
			command: "poll",
		},
	}
}
