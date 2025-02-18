package command

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	handlers "github.com/holeyko/innany-tgbot/internal/inanny/handlers"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/polls"
)

type CommandHandler struct {
	command string
}

func (handler CommandHandler) IsSutable(command *string) bool {
	return *command == handler.command
}

var commandHandlers = [...]handlers.TgUpdateHandler[string]{
	NewPollCommandHandler(),
	NewBinPollCommandHandler(),
	NewHelloCommandHandler(),
}

func FindCommandHandler(command string) handlers.TgUpdateHandler[string] {
	for _, handler := range commandHandlers {
		if handler.IsSutable(&command) {
			return handler
		}
	}

	return nil
}

func applyFlagsToPollConfig(pollConfig *tgbot.SendPollConfig, flags []polls.Flag) {
	pollConfig.IsAnonymous = false
	for _, flag := range flags {
		switch flag {
		case polls.Anonymous:
			pollConfig.IsAnonymous = true
		case polls.Multipoll:
			pollConfig.AllowsMultipleAnswers = true
		}
	}
}
