package main

import (
	"log"
	"os"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	commands "github.com/holeyko/innany-tgbot/internal/inanny/handlers/commands"
)

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Can't find TELEGRAM_BOT_TOKEN environment variable")
	}

	bot, err := tgbot.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Telegram bot Innany was started")
	handeRequests(bot)
}

func buildUpdateConfig() tgbot.UpdateConfig {
	updateConfig := tgbot.NewUpdate(0)
	updateConfig.Timeout = 30

	return updateConfig
}

func handeRequests(bot *tgbot.BotAPI) {
	updates := bot.GetUpdatesChan(buildUpdateConfig())

	for update := range updates {
		if message := update.Message; message != nil {
			if command := message.Command(); command != "" {
				if handler := commands.FindCommandHandler(command); handler != nil {
					err := handler.Handle(bot, &update)
					if err != nil {
						sendErrorResponse(bot, message.Chat.ID, err)
					}
				}
			}
		}
	}
}

func sendErrorResponse(bot *tgbot.BotAPI, chatId int64, err error) {
	messageConfig := tgbot.NewMessage(chatId, err.Error())
	bot.Send(messageConfig)
}
