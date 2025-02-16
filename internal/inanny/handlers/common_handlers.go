package handlers

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgUpdateHandler[T any] interface {
	Handle(bot *tgbot.BotAPI, update *tgbot.Update) error
	IsSutable(value *T) bool
}
