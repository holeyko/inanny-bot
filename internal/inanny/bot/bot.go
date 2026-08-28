package bot

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	commands "github.com/holeyko/innany-tgbot/internal/inanny/bot/handlers/commands"
	customcommands "github.com/holeyko/innany-tgbot/internal/inanny/features/customcommands"
	messages "github.com/holeyko/innany-tgbot/internal/inanny/features/messages"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/features/polls"
)

var messageRepository = messages.NewRepository()

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
	customcommands.StartDraftCleanup()
	debugLog("custom command draft cleanup started")
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
		if err := storeMessage(message); err != nil {
			return err
		}

		if handled, commandDraftErr := commands.TryHandleCustomCommandDraftReply(bot, update); handled || commandDraftErr != nil {
			return commandDraftErr
		}

		if handled, draftErr := polls.TryHandlePollDraftReply(bot, update); handled || draftErr != nil {
			return draftErr
		}

		if command := message.Command(); command != "" {
			err = handleCommand(bot, update)
		} else if handled, mentionErr := commands.TryHandleCrewMention(bot, update); handled {
			err = mentionErr
		}
	}

	return
}

func storeMessage(message *tgbot.Message) error {
	text := message.Text
	if text == "" {
		text = message.Caption
	}
	if text == "" || message.Chat == nil || message.MessageID == 0 {
		return nil
	}

	sender := ""
	if message.From != nil {
		sender = message.From.UserName
		if sender == "" {
			sender = strings.TrimSpace(strings.Join([]string{message.From.FirstName, message.From.LastName}, " "))
		}
	}
	if sender == "" && message.SenderChat != nil {
		sender = message.SenderChat.Title
	}

	sentAt := time.Now()
	if message.Date != 0 {
		sentAt = time.Unix(int64(message.Date), 0)
	}

	replyToMessageID := int64(0)
	if message.ReplyToMessage != nil {
		replyToMessageID = int64(message.ReplyToMessage.MessageID)
	}

	return messageRepository.Save(messages.Message{
		ChatID:           message.Chat.ID,
		MessageID:        int64(message.MessageID),
		Sender:           sender,
		Text:             text,
		ReplyToMessageID: replyToMessageID,
		SentAt:           sentAt,
	})
}

func tryHandleCallback(bot *tgbot.BotAPI, update *tgbot.Update) (err error) {
	if callback := update.CallbackQuery; callback != nil {
		handled, callbackErr := commands.HandleHelpCallback(bot, update)
		if handled {
			return callbackErr
		}

		log.Println("Callback data:", callback.Data)
	}

	return
}

func handleCommand(bot *tgbot.BotAPI, update *tgbot.Update) (err error) {
	message := update.Message
	command := message.Command()
	debugLog("handling command %q for chat_id=%d", command, message.Chat.ID)

	if handler := commands.FindCommandHandler(command); handler != nil {
		err = handler.Handle(bot, update)
	} else if handled, customErr := commands.HandleCustomCommand(bot, update, command); handled {
		err = customErr
	} else {
		err = fmt.Errorf("Command %s doesn't exist", command)
	}

	return
}

func handleError(bot *tgbot.BotAPI, update *tgbot.Update, err error) {
	if update.Message == nil {
		log.Println("Error while handling update without message:", err)
	} else {
		sendErrorResponse(bot, update, err)
	}
}

func sendErrorResponse(bot *tgbot.BotAPI, update *tgbot.Update, err error) {
	message := update.Message
	replyToMessageID := message.MessageID
	if message.ReplyToMessage != nil {
		replyToMessageID = message.ReplyToMessage.MessageID
	}

	messageConfig := tgbot.NewMessage(message.Chat.ID, err.Error())
	messageConfig.ReplyToMessageID = replyToMessageID
	bot.Send(messageConfig)
}

func debugLog(format string, args ...any) {
	if strings.EqualFold(os.Getenv("DEBUG"), "true") {
		log.Printf("DEBUG: "+format, args...)
	}
}
