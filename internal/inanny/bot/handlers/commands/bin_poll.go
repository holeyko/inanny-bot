package command

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	polls "github.com/holeyko/innany-tgbot/internal/inanny/features/polls"
)

type BinPollCommandHandler struct {
	CommandHandler
}

func (handler BinPollCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	command, err := polls.ParseBinPollCommand(update.Message.CommandArguments())
	if err != nil {
		return err
	}

	user, err := polls.UpsertUser(polls.UserDto{
		TelegramLogin: update.Message.From.UserName,
		FirstName:     update.Message.From.FirstName,
		LastName:      update.Message.From.LastName,
	})
	if err != nil {
		return err
	}

	if len(command.PendingFlags) == 0 {
		return polls.SendPoll(bot, &command.Poll, update.Message)
	}

	draft, err := polls.CreatePollDraft(polls.CreatePollDraftDto{
		ChatID:          update.Message.Chat.ID,
		UserID:          user.ID,
		Command:         command.Poll.Command,
		Title:           command.Poll.Title,
		PendingFlags:    command.PendingFlags,
		SourceMessageID: int64(update.Message.MessageID),
	})
	if err != nil {
		return err
	}

	prompt, err := polls.SendDraftPrompt(bot, update.Message.Chat.ID, update.Message.MessageID, polls.DraftStepPrompt(draft.PendingFlags[0]))
	if err != nil {
		_ = polls.DeletePollDraftByID(draft.ID)
		return err
	}

	_, err = polls.UpdatePollDraft(polls.UpdatePollDraftDto{
		ID:              draft.ID,
		PendingFlags:    draft.PendingFlags,
		PromptMessageID: int64(prompt.MessageID),
	})
	if err != nil {
		_ = polls.DeletePollDraftByID(draft.ID)
		return err
	}

	return nil
}

func NewBinPollCommandHandler() BinPollCommandHandler {
	return BinPollCommandHandler{
		CommandHandler: CommandHandler{
			command: "bin_poll",
		},
	}
}
