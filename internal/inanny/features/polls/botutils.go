package poll

import (
	"encoding/json"
	"errors"
	"slices"
	"strconv"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendPoll(bot *tgbot.BotAPI, poll *Poll, message *tgbot.Message) (err error) {
	err = sendPollToChat(bot, poll, message.Chat.ID, message)
	return
}

func SendPollToChat(bot *tgbot.BotAPI, poll *Poll, chatID int64) (err error) {
	err = sendPollToChat(bot, poll, chatID, nil)
	return
}

func sendPollToChat(bot *tgbot.BotAPI, poll *Poll, chatID int64, sourceMessage *tgbot.Message) (err error) {
	err = CheckPoll(poll)
	if err != nil {
		return
	}

	var pollMessage tgbot.Message
	if slices.Contains(poll.Flags, Ext) {
		pollMessage, err = sendExtendablePoll(bot, poll, chatID)
	} else {
		pollConfig := tgbot.NewPoll(
			chatID,
			poll.Title,
			poll.Options...,
		)

		applyFlagsToPollConfig(&pollConfig, poll.Flags)
		pollMessage, err = bot.Send(pollConfig)
	}

	if err != nil {
		return
	}

	err = postPollProcessing(bot, poll, sourceMessage, &pollMessage)
	return
}

func CheckPoll(poll *Poll) (err error) {
	if len(poll.Options) < 2 {
		err = errors.New("Should be at least 2 options")
	} else if slices.Contains(poll.Flags, Ext) && slices.Contains(poll.Flags, Anonymous) {
		err = errors.New("Extendable polls cannot be anonymous")
	}

	return
}

func sendExtendablePoll(bot *tgbot.BotAPI, poll *Poll, chatID int64) (tgbot.Message, error) {
	params := tgbot.Params{
		"chat_id":                 strconv.FormatInt(chatID, 10),
		"question":                poll.Title,
		"is_anonymous":            "false",
		"allows_multiple_answers": strconv.FormatBool(slices.Contains(poll.Flags, Multipoll)),
		"allow_adding_options":    "true",
	}
	if err := params.AddInterface("options", poll.Options); err != nil {
		return tgbot.Message{}, err
	}

	response, err := bot.MakeRequest("sendPoll", params)
	if err != nil {
		return tgbot.Message{}, err
	}

	var message tgbot.Message
	if err := json.Unmarshal(response.Result, &message); err != nil {
		return tgbot.Message{}, err
	}

	return message, nil
}

func applyFlagsToPollConfig(pollConfig *tgbot.SendPollConfig, flags []Flag) {
	pollConfig.IsAnonymous = false
	for _, flag := range flags {
		switch flag {
		case Anonymous:
			pollConfig.IsAnonymous = true
		case Multipoll:
			pollConfig.AllowsMultipleAnswers = true
		}
	}
}

func postPollProcessing(bot *tgbot.BotAPI, poll *Poll, message *tgbot.Message, pollMessage *tgbot.Message) (err error) {
	if slices.Contains(poll.Flags, Pin) {
		_, err = pinMessage(bot, pollMessage.Chat.ID, pollMessage.MessageID, true)
	}
	if sourceMessage := message; sourceMessage != nil && slices.Contains(poll.Flags, Remove) {
		_, err = removeMessage(bot, message.Chat.ID, message.MessageID)
	}

	return
}

func pinMessage(
	bot *tgbot.BotAPI,
	chatId int64,
	messageId int,
	notify bool,
) (response *tgbot.APIResponse, err error) {
	pinConfig := tgbot.PinChatMessageConfig{
		ChatID:              chatId,
		MessageID:           messageId,
		DisableNotification: !notify,
	}

	response, err = bot.Request(pinConfig)
	return
}

func removeMessage(bot *tgbot.BotAPI, chatId int64, messageId int) (response *tgbot.APIResponse, err error) {
	deleteMessge := tgbot.NewDeleteMessage(chatId, messageId)
	_, err = bot.Request(deleteMessge)
	return
}
