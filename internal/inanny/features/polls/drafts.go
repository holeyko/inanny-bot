package poll

import (
	ctx "context"
	errors2 "errors"
	fmt2 "fmt"
	log2 "log"
	strings2 "strings"
	time2 "time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/holeyko/innany-tgbot/internal/generated/queries"
	"github.com/holeyko/innany-tgbot/internal/inanny/infra/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const draftTTL = time2.Hour

var ErrPollDraftNotFound = errors2.New("poll draft not found")

type PollDraft struct {
	ID              int64
	ChatID          int64
	UserID          int64
	Command         string
	Title           string
	Options         []string
	Flags           []Flag
	PinEnabled      bool
	CronExpr        string
	StepIndex       int32
	SourceMessageID int64
	PromptMessageID int64
	CreatedAt       time2.Time
	UpdatedAt       time2.Time
}

type CreatePollDraftDto struct {
	ChatID          int64
	UserID          int64
	Command         string
	Title           string
	Options         []string
	Flags           []Flag
	PinEnabled      bool
	CronExpr        string
	StepIndex       int32
	SourceMessageID int64
	PromptMessageID int64
}

type UpdatePollDraftDto struct {
	ID              int64
	PinEnabled      bool
	CronExpr        string
	StepIndex       int32
	PromptMessageID int64
}

func DraftInteractiveSteps(flags []Flag) []Flag {
	steps := make([]Flag, 0, len(flags))
	for _, flag := range flags {
		switch flag {
		case Cron:
			steps = append(steps, flag)
		}
	}

	return steps
}

func DraftStepPrompt(step Flag) string {
	switch step {
	case Cron:
		return "Send cron expression on reply this message"
	default:
		return ""
	}
}

func ParseDraftStepAnswer(step Flag, answer string) (bool, string, error) {
	answer = strings2.TrimSpace(strings2.ToLower(answer))

	switch step {
	case Cron:
		if answer == "" {
			return false, "", errors2.New("Cron expression can't be empty")
		}
		if err := ValidateCronExpr(answer); err != nil {
			return false, "", err
		}
		return false, answer, nil
	default:
		return false, "", fmt2.Errorf("Unsupported step: %s", step)
	}
}

func BuildFinalPoll(draft *PollDraft) Poll {
	flags := make([]Flag, 0, len(draft.Flags))
	for _, flag := range draft.Flags {
		switch flag {
		case Cron:
			continue
		case Pin:
			flags = append(flags, flag)
		default:
			flags = append(flags, flag)
		}
	}

	return Poll{
		Command: draft.Command,
		Title:   draft.Title,
		Options: draft.Options,
		Flags:   flags,
	}
}

func StartPollFlow(bot *tgbot.BotAPI, update *tgbot.Update, poll Poll) error {
	user, err := UpsertUser(UserDto{
		TelegramLogin: update.Message.From.UserName,
		FirstName:     update.Message.From.FirstName,
		LastName:      update.Message.From.LastName,
	})
	if err != nil {
		return err
	}

	steps := DraftInteractiveSteps(poll.Flags)
	if len(steps) == 0 {
		return createPollFromFinalState(bot, update.Message, user.ID, &PollDraft{
			ChatID:    update.Message.Chat.ID,
			UserID:    user.ID,
			Command:   poll.Command,
			Title:     poll.Title,
			Options:   poll.Options,
			Flags:     poll.Flags,
			StepIndex: 0,
		})
	}

	draft, err := CreatePollDraft(CreatePollDraftDto{
		ChatID:          update.Message.Chat.ID,
		UserID:          user.ID,
		Command:         poll.Command,
		Title:           poll.Title,
		Options:         poll.Options,
		Flags:           poll.Flags,
		SourceMessageID: int64(update.Message.MessageID),
		StepIndex:       0,
	})
	if err != nil {
		return err
	}

	prompt, err := SendDraftPrompt(bot, update.Message.Chat.ID, update.Message.MessageID, DraftStepPrompt(steps[0]))
	if err != nil {
		_ = DeletePollDraftByID(draft.ID)
		return err
	}

	_, err = UpdatePollDraft(UpdatePollDraftDto{
		ID:              draft.ID,
		PinEnabled:      draft.PinEnabled,
		CronExpr:        draft.CronExpr,
		StepIndex:       0,
		PromptMessageID: int64(prompt.MessageID),
	})
	if err != nil {
		_ = DeletePollDraftByID(draft.ID)
		return err
	}

	return nil
}

func createPollFromFinalState(bot *tgbot.BotAPI, message *tgbot.Message, userID int64, draft *PollDraft) error {
	finalPoll := BuildFinalPoll(draft)
	if draft.CronExpr == "" {
		return SendPoll(bot, &finalPoll, message)
	}

	if err := CheckPoll(&finalPoll); err != nil {
		return err
	}

	storedPoll, err := CreateStoredPoll(CreateStoredPollDto{
		ChatID:   draft.ChatID,
		UserID:   userID,
		Command:  draft.Command,
		Poll:     finalPoll,
		CronExpr: draft.CronExpr,
	})
	if err != nil {
		return err
	}

	if err := RegisterCronPoll(storedPoll); err != nil {
		return err
	}

	_, err = SendDraftPrompt(bot, message.Chat.ID, message.MessageID, fmt2.Sprintf("Cron poll #%d was created", storedPoll.ID))
	return err
}

func GetPollDraftByPromptMessageID(chatID int64, userID int64, promptMessageID int64) (*PollDraft, error) {
	return db.Execute(func(q *queries.Queries) (*PollDraft, error) {
		draft, err := q.GetPollDraftByPromptMessageID(ctx.Background(), queries.GetPollDraftByPromptMessageIDParams{
			ChatID:          chatID,
			UserID:          userID,
			PromptMessageID: promptMessageID,
		})
		if err != nil {
			if errors2.Is(err, pgx.ErrNoRows) {
				return nil, ErrPollDraftNotFound
			}
			return nil, err
		}

		return mapPollDraft(&draft), nil
	})
}

func CreatePollDraft(dto CreatePollDraftDto) (*PollDraft, error) {
	return db.Execute(func(q *queries.Queries) (*PollDraft, error) {
		draft, err := q.CreatePollDraft(ctx.Background(), queries.CreatePollDraftParams{
			ChatID:          dto.ChatID,
			UserID:          dto.UserID,
			Command:         dto.Command,
			Title:           dto.Title,
			Options:         dto.Options,
			Flags:           FlagsToStrings(dto.Flags),
			PinEnabled:      dto.PinEnabled,
			CronExpr:        pgtype.Text{String: dto.CronExpr, Valid: dto.CronExpr != ""},
			StepIndex:       dto.StepIndex,
			SourceMessageID: dto.SourceMessageID,
			PromptMessageID: dto.PromptMessageID,
		})
		if err != nil {
			return nil, err
		}

		return mapPollDraft(&draft), nil
	})
}

func UpdatePollDraft(dto UpdatePollDraftDto) (*PollDraft, error) {
	return db.Execute(func(q *queries.Queries) (*PollDraft, error) {
		draft, err := q.UpdatePollDraft(ctx.Background(), queries.UpdatePollDraftParams{
			ID:              dto.ID,
			PinEnabled:      dto.PinEnabled,
			CronExpr:        pgtype.Text{String: dto.CronExpr, Valid: dto.CronExpr != ""},
			StepIndex:       dto.StepIndex,
			PromptMessageID: dto.PromptMessageID,
		})
		if err != nil {
			return nil, err
		}

		return mapPollDraft(&draft), nil
	})
}

func DeletePollDraftByID(id int64) error {
	_, err := db.Execute(func(q *queries.Queries) (struct{}, error) {
		_, err := q.DeletePollDraftByID(ctx.Background(), id)
		return struct{}{}, err
	})
	return err
}

func CleanupExpiredPollDrafts() (int64, error) {
	return db.Execute(func(q *queries.Queries) (int64, error) {
		return q.DeleteExpiredPollDrafts(ctx.Background())
	})
}

func StartDraftCleanup() {
	if _, err := CleanupExpiredPollDrafts(); err != nil {
		log2.Println("Error while cleaning expired poll drafts:", err)
	}

	go func() {
		ticker := time2.NewTicker(10 * time2.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if _, err := CleanupExpiredPollDrafts(); err != nil {
				log2.Println("Error while cleaning expired poll drafts:", err)
			}
		}
	}()
}

func SendDraftPrompt(bot *tgbot.BotAPI, chatID int64, replyToMessageID int, text string) (*tgbot.Message, error) {
	messageConfig := tgbot.NewMessage(chatID, text)
	messageConfig.ReplyToMessageID = replyToMessageID
	message, err := bot.Send(messageConfig)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func TryHandlePollDraftReply(bot *tgbot.BotAPI, update *tgbot.Update) (bool, error) {
	message := update.Message
	if message == nil || message.From == nil || message.ReplyToMessage == nil {
		return false, nil
	}

	user, err := UpsertUser(UserDto{
		TelegramLogin: message.From.UserName,
		FirstName:     message.From.FirstName,
		LastName:      message.From.LastName,
	})
	if err != nil {
		return false, err
	}

	draft, err := GetPollDraftByPromptMessageID(message.Chat.ID, user.ID, int64(message.ReplyToMessage.MessageID))
	if err != nil {
		if errors2.Is(err, ErrPollDraftNotFound) {
			return false, nil
		}
		return true, err
	}

	if draft.IsExpired() {
		if err := DeletePollDraftByID(draft.ID); err != nil {
			return true, err
		}
		return true, errors2.New("Poll draft expired, start again")
	}

	steps := DraftInteractiveSteps(draft.Flags)
	if int(draft.StepIndex) >= len(steps) {
		if err := DeletePollDraftByID(draft.ID); err != nil {
			return true, err
		}
		return true, errors2.New("Poll draft is already completed")
	}

	step := steps[draft.StepIndex]
	pinEnabled := draft.PinEnabled
	cronExpr := draft.CronExpr
	stepPinEnabled, stepCronExpr, err := ParseDraftStepAnswer(step, message.Text)
	if err != nil {
		return true, err
	}

	switch step {
	case Pin:
		pinEnabled = stepPinEnabled
	case Cron:
		cronExpr = stepCronExpr
	}

	nextStepIndex := draft.StepIndex + 1
	if int(nextStepIndex) < len(steps) {
		nextPrompt, err := SendDraftPrompt(bot, message.Chat.ID, message.MessageID, DraftStepPrompt(steps[nextStepIndex]))
		if err != nil {
			return true, err
		}

		_, err = UpdatePollDraft(UpdatePollDraftDto{
			ID:              draft.ID,
			PinEnabled:      pinEnabled,
			CronExpr:        cronExpr,
			StepIndex:       nextStepIndex,
			PromptMessageID: int64(nextPrompt.MessageID),
		})
		if err != nil {
			_ = DeletePollDraftByID(draft.ID)
			return true, err
		}

		return true, nil
	}

	draft.PinEnabled = pinEnabled
	draft.CronExpr = cronExpr
	draft.StepIndex = nextStepIndex

	if err := createPollFromFinalState(bot, message, user.ID, draft); err != nil {
		return true, err
	}

	if err := DeletePollDraftByID(draft.ID); err != nil {
		return true, err
	}

	return true, nil
}

func mapPollDraft(dbDraft *queries.PollDraft) *PollDraft {
	cronExpr := ""
	if dbDraft.CronExpr.Valid {
		cronExpr = dbDraft.CronExpr.String
	}

	return &PollDraft{
		ID:              dbDraft.ID,
		ChatID:          dbDraft.ChatID,
		UserID:          dbDraft.UserID,
		Command:         dbDraft.Command,
		Title:           dbDraft.Title,
		Options:         dbDraft.Options,
		Flags:           StringsToFlags(dbDraft.Flags),
		PinEnabled:      dbDraft.PinEnabled,
		CronExpr:        cronExpr,
		StepIndex:       dbDraft.StepIndex,
		SourceMessageID: dbDraft.SourceMessageID,
		PromptMessageID: dbDraft.PromptMessageID,
		CreatedAt:       dbDraft.CreatedAt.Time,
		UpdatedAt:       dbDraft.UpdatedAt.Time,
	}
}

func (draft *PollDraft) IsExpired() bool {
	return time2.Since(draft.CreatedAt) > draftTTL
}
