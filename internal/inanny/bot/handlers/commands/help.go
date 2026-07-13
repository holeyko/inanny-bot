package command

import (
	strings "strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HelpCommandHandler struct {
	CommandHandler
}

func (handler HelpCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	lines := []string{
		"Innany commands",
		"",
		"Common:",
		"/help - show this help",
		"/hello - say hello",
		"",
		"Custom commands:",
		"/commands - list custom commands in this chat",
		"/commands create - create a custom command from a reply template",
		"/commands delete <id> - delete custom command as creator or chat admin",
		"Template: <name> <existing_command> <body>",
		"Parameters: use $1, $2, ... in body and pass values as /name value1 value2",
		"Example: lunch bp [pin, rm] Lunch at $1 p.m.?",
		"",
		"Polls:",
		"/poll [flags] Title - custom poll; options go on next lines",
		"/bp [flags] Title - poll with Да/Нет",
		"/tp [flags] Title - poll with Да/Нет/Тык",
		"/polls - list your cron polls in this chat",
		"/polls delete <id> - delete cron poll",
		"",
		"Poll flags:",
		"- ano: anonymous",
		"- mul: multi-answer",
		"- pin: pin poll",
		"- rm: remove command message",
		"- cron: schedule poll",
	}

	return sendReply(bot, update, strings.Join(lines, "\n"))
}

func NewHelpCommandHandler() HelpCommandHandler {
	return HelpCommandHandler{
		CommandHandler: CommandHandler{command: "help"},
	}
}
