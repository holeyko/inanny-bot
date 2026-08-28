package command

import (
	strings "strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	helpCallbackPrefix   = "help:"
	helpCallbackMain     = helpCallbackPrefix + "main"
	helpCallbackPolls    = helpCallbackPrefix + "polls"
	helpCallbackCommands = helpCallbackPrefix + "commands"
	helpCallbackCrews    = helpCallbackPrefix + "crews"
	helpCallbackAI       = helpCallbackPrefix + "ai"
)

type HelpCommandHandler struct {
	CommandHandler
}

func (handler HelpCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	messageConfig := tgbot.NewMessage(update.Message.Chat.ID, helpText(helpCallbackMain))
	messageConfig.ReplyToMessageID = update.Message.MessageID
	messageConfig.ReplyMarkup = helpKeyboard(helpCallbackMain)

	_, err := bot.Send(messageConfig)
	return err
}

func HandleHelpCallback(bot *tgbot.BotAPI, update *tgbot.Update) (bool, error) {
	callback := update.CallbackQuery
	if callback == nil || !strings.HasPrefix(callback.Data, helpCallbackPrefix) {
		return false, nil
	}

	if _, err := bot.Request(tgbot.NewCallback(callback.ID, "")); err != nil {
		return true, err
	}

	text := helpText(callback.Data)
	if text == "" || callback.Message == nil {
		return true, nil
	}

	messageConfig := tgbot.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		helpKeyboard(callback.Data),
	)
	_, err := bot.Send(messageConfig)
	return true, err
}

func helpText(section string) string {
	var lines []string
	switch section {
	case helpCallbackPolls:
		lines = []string{
			"Polls",
			"/poll, /p [flags] Title - custom poll; add options on following lines",
			"/bp, /bin_poll [flags] Title - poll with Да/Нет",
			"/tp, /tripple_poll [flags] Title - poll with Да/Нет/Тык",
			"/polls - list scheduled polls; /polls delete <id> - delete one",
			"",
			"Flags:",
			"ano - anonymous",
			"mul - allow multiple answers",
			"pin - pin the poll",
			"rm - remove the command message",
			"cron - schedule the poll",
			"ext - let users add options; not anonymous",
			"",
			"Example: /poll [mul, pin] Lunch?\nPizza\nSushi",
		}
	case helpCallbackCommands:
		lines = []string{
			"Commands",
			"/commands - list custom commands in this chat",
			"/commands create - create one from a reply template",
			"/commands delete <id> - delete one as creator or chat admin",
			"",
			"Flags: none.",
			"Template: <name> <existing_command> [body]; use $1, $2, ... for values",
			"Example: lunch bp [pin, rm] Lunch at $1 p.m.?",
		}
	case helpCallbackCrews:
		lines = []string{
			"Crews",
			"/crew <name> - mention all members of a crew",
			"@<name> - mention all members in a regular chat message",
			"/crew create <name> <username1> <username2> ... - create a crew",
			"/crew delete <name> - delete a crew",
			"/crew <name> add <username1> <username2> ... - add members",
			"/crew <name> delete <username> - remove a member",
			"",
			"Crew names support letters from any language and can be up to 128 characters.",
			"Use letters, numbers, or underscores in crew names.",
			"Separate usernames with spaces, commas, or both.",
			"The creator is added automatically; any number of additional members may be provided.",
			"The final crew must have at least 2 unique members. The creator may also be listed explicitly.",
			"Examples: @alice sends /crew create developers bob, or /crew create developers alice bob",
		}
	case helpCallbackAI:
		lines = []string{
			"AI features",
			"/summarize, /summarise, /sum - summarize a replied conversation",
			"",
			"Flags: none.",
			"Uses up to the last 10,000 words and follows the conversation's main language.",
			"Example: reply to the first message, then send /sum",
		}
	case helpCallbackMain:
		lines = []string{
			"Innany commands",
			"/help, /man - show this help",
			"/hello - say hello",
			"/summarize, /summarise, /sum - summarize a replied conversation",
			"/commands - manage custom commands",
			"/crew - manage and mention crews",
			"/poll, /p - create a custom poll",
			"/bin_poll, /bp - create a Да/Нет poll",
			"/tripple_poll, /tp - create a Да/Нет/Тык poll",
			"/polls - manage scheduled polls",
		}
	default:
		return ""
	}

	return strings.Join(lines, "\n")
}

func helpKeyboard(section string) tgbot.InlineKeyboardMarkup {
	if section == helpCallbackMain {
		return tgbot.NewInlineKeyboardMarkup(
			tgbot.NewInlineKeyboardRow(
				tgbot.NewInlineKeyboardButtonData("Polls", helpCallbackPolls),
				tgbot.NewInlineKeyboardButtonData("Commands", helpCallbackCommands),
				tgbot.NewInlineKeyboardButtonData("Crews", helpCallbackCrews),
				tgbot.NewInlineKeyboardButtonData("AI features", helpCallbackAI),
			),
		)
	}

	return tgbot.NewInlineKeyboardMarkup(
		tgbot.NewInlineKeyboardRow(
			tgbot.NewInlineKeyboardButtonData("Back to help", helpCallbackMain),
		),
	)
}

func NewHelpCommandHandler() HelpCommandHandler {
	return HelpCommandHandler{
		CommandHandler: CommandHandler{command: "help", aliases: []string{"man"}},
	}
}
