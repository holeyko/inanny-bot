package bot

import (
	"fmt"
	"log"
	"os"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	commands "github.com/holeyko/innany-tgbot/internal/inanny/bot/handlers/commands"
)

func StartBot() {
	bot := createBot()
	startHandeRequests(bot)
}

func createBot() *tgbot.BotAPI {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Can't find TELEGRAM_BOT_TOKEN environment variable")
	}

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

	for update := range updates {
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
