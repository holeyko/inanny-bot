package command

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/holeyko/innany-tgbot/internal/inanny/features/ai"
)

type AICommandHandler struct {
	CommandHandler
	router *ai.Router
}

func NewAICommandHandler(router *ai.Router) AICommandHandler {
	return AICommandHandler{CommandHandler: CommandHandler{command: "ai"}, router: router}
}

func (handler AICommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	if update.Message == nil {
		return nil
	}
	args := strings.Fields(update.Message.CommandArguments())
	if len(args) != 1 || args[0] != "quota" {
		return sendReply(bot, update, "Usage: /ai quota — view this chat's remaining daily DeepSeek Flash requests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	quota, err := handler.router.Usage(ctx, update.Message.Chat.ID)
	if err != nil {
		return errors.New("I couldn't read the AI quota. Please try again later")
	}
	text := fmt.Sprintf("DeepSeek Flash daily quota for this chat\nRemaining: %d/%d\nUsed: %d\nResets: %s UTC\nAttempts count even if DeepSeek fails. After the quota is used, the default provider is used. /ai quota does not consume requests.",
		quota.Limit-quota.Used, quota.Limit, quota.Used, quota.ResetsAt.UTC().Format("2006-01-02 15:04"))
	if !handler.router.PrimaryConfigured() {
		text += "\nOpenRouter is not configured; the default provider is currently used."
	}
	return sendReply(bot, update, text)
}
