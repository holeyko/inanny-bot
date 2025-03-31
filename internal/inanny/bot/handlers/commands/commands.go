package command

import (
	handlers "github.com/holeyko/innany-tgbot/internal/inanny/bot/handlers"
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
