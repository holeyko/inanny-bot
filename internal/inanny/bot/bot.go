package bot

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	commands "github.com/holeyko/innany-tgbot/internal/inanny/bot/handlers/commands"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/features/polls"
)

func StartBot() {
	debugLog("starting bot bootstrap")
	bot := createBot()
	debugLog("telegram bot client created for account %q", bot.Self.UserName)
	if err := polls.StartScheduler(bot); err != nil {
		log.Println("Cron poll scheduler started without persisted polls:", err)
	} else {
		debugLog("cron poll scheduler started")
	}
	polls.StartDraftCleanup()
	debugLog("poll draft cleanup started")
	startHandeRequests(bot)
}

func createBot() *tgbot.BotAPI {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Can't find TELEGRAM_BOT_TOKEN environment variable")
	}
	debugLog("required environment variables present: TELEGRAM_BOT_TOKEN=%t DB_HOST=%t DB_USER=%t DB_NAME=%t", botToken != "", os.Getenv("DB_HOST") != "", os.Getenv("DB_USER") != "", os.Getenv("DB_NAME") != "")

	bot, err := tgbot.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	return bot
}

func buildUpdateConfig() tgbot.UpdateConfig {
	updateConfig := tgbot.NewUpdate(0)
	updateConfig.Timeout = 30

	return updateConfig
}

func startHandeRequests(bot *tgbot.BotAPI) {
	updates := bot.GetUpdatesChan(buildUpdateConfig())
	log.Println("Telegram bot Innany was started")
	debugLog("telegram update polling started")

	for update := range updates {
		debugLog("received update id=%d has_message=%t has_callback=%t", update.UpdateID, update.Message != nil, update.CallbackQuery != nil)
		go handleRequest(bot, &update)
	}
}

func handleRequest(bot *tgbot.BotAPI, update *tgbot.Update) {
	var err error

	if err == nil {
		err = tryHandleMessage(bot, update)
	}
	if err == nil {
		err = tryHandleCallback(bot, update)
	}

	if err != nil {
		handleError(bot, update, err)
	}
}

func tryHandleMessage(bot *tgbot.BotAPI, update *tgbot.Update) (err error) {
	if message := update.Message; message != nil {
		if handled, draftErr := polls.TryHandlePollDraftReply(bot, update); handled || draftErr != nil {
			return draftErr
		}

		if command := message.Command(); command != "" {
			err = handleCommand(bot, update)
		}
	}

	return
}

func tryHandleCallback(bot *tgbot.BotAPI, update *tgbot.Update) (err error) {
	if callback := update.CallbackQuery; callback != nil {
		callbackData := callback.Data
		log.Println("Callback data:", callbackData)
	}

	return
}

func handleCommand(bot *tgbot.BotAPI, update *tgbot.Update) (err error) {
	message := update.Message
	command := message.Command()
	debugLog("handling command %q for chat_id=%d", command, message.Chat.ID)

	if handler := commands.FindCommandHandler(command); handler != nil {
		err = handler.Handle(bot, update)
	} else {
		err = fmt.Errorf("Command %s doesn't exist", command)
	}

	return
}

func handleError(bot *tgbot.BotAPI, update *tgbot.Update, err error) {
	if update.Message == nil {
		log.Println("Error while handling update without message:", err)
	} else {
		sendErrorResponse(bot, update.Message.Chat.ID, err)
	}
}

func sendErrorResponse(bot *tgbot.BotAPI, chatId int64, err error) {
	messageConfig := tgbot.NewMessage(chatId, err.Error())
	bot.Send(messageConfig)
}

func debugLog(format string, args ...any) {
	if strings.EqualFold(os.Getenv("DEBUG"), "true") {
		log.Printf("DEBUG: "+format, args...)
	}
}
