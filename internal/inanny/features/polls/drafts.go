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

const (
	binPollCommandName = "bin_poll"
	draftTTL           = time2.Hour
)

var (
	binPollOptions       = []string{"Да", "Нет"}
	ErrPollDraftNotFound = errors2.New("poll draft not found")
)

type BinPollCommand struct {
	Poll         Poll
	PendingFlags []Flag
}

type PollDraft struct {
	ID              int64
	ChatID          int64
	UserID          int64
	Command         string
	Title           string
	PendingFlags    []Flag
	PinEnabled      bool
	CronExpr        string
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
	PendingFlags    []Flag
	PinEnabled      bool
	CronExpr        string
	SourceMessageID int64
	PromptMessageID int64
}

type UpdatePollDraftDto struct {
	ID              int64
	PendingFlags    []Flag
	PinEnabled      bool
	CronExpr        string
	PromptMessageID int64
}

func ParseBinPollCommand(input string) (BinPollCommand, error) {
	flags, body, err := parseBinPollCommandFlagsAndBody(input)
	if err != nil {
		return BinPollCommand{}, err
	}

	if strings2.TrimSpace(body) == "" {
		return BinPollCommand{}, errors2.New("Body can't be empty")
	}

	pendingFlags := make([]Flag, 0, len(flags))
	seen := map[Flag]struct{}{}
	for _, flag := range flags {
		switch flag {
		case Pin, Cron:
			if _, ok := seen[flag]; ok {
				return BinPollCommand{}, fmt2.Errorf("Duplicate flag: %s", flag)
			}
			seen[flag] = struct{}{}
			pendingFlags = append(pendingFlags, flag)
		default:
			return BinPollCommand{}, fmt2.Errorf("Unsupported bin_poll flag: %s", flag)
		}
	}

	return BinPollCommand{
		Poll: Poll{
			Command: binPollCommandName,
			Title:   strings2.TrimSpace(body),
			Options: binPollOptions,
		},
		PendingFlags: pendingFlags,
	}, nil
}

func parseBinPollCommandFlagsAndBody(input string) ([]Flag, string, error) {
	input = strings2.TrimSpace(input)
	if input == "" {
		return nil, "", errors2.New("Body can't be empty")
	}

	if input[0] != '[' {
		return nil, input, nil
	}

	closeIndex := strings2.Index(input, "]")
	if closeIndex == -1 {
		return nil, "", errors2.New("Flags should be closed with ]")
	}

	flagsPart := strings2.TrimSpace(input[1:closeIndex])
	body := strings2.TrimSpace(input[closeIndex+1:])
	if body == "" {
		return nil, "", errors2.New("Body can't be empty")
	}

	if flagsPart == "" {
		return nil, body, nil
	}

	rawFlags := strings2.Split(flagsPart, ",")
	flags := make([]Flag, 0, len(rawFlags))
	for _, rawFlag := range rawFlags {
		flag := Flag(strings2.ToLower(strings2.TrimSpace(rawFlag)))
		if flag == "" {
			return nil, "", errors2.New("Flag can't be empty")
		}
		flags = append(flags, flag)
	}

	return flags, body, nil
}

func DraftStepPrompt(step Flag) string {
	switch step {
	case Pin:
		return "Закрепить опрос? Ответь Да или Нет."
	case Cron:
		return "Пришли cron expression."
	default:
		return ""
	}
}

func ParseDraftStepAnswer(step Flag, answer string) (bool, string, error) {
	answer = strings2.TrimSpace(strings2.ToLower(answer))

	switch step {
	case Pin:
		switch answer {
		case "да":
			return true, "", nil
		case "нет":
			return false, "", nil
		default:
			return false, "", errors2.New("Ответь Да или Нет")
		}
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
	flags := make([]Flag, 0, 1)
	if draft.PinEnabled {
		flags = append(flags, Pin)
	}

	return Poll{
		Command: draft.Command,
		Title:   draft.Title,
		Options: binPollOptions,
		Flags:   flags,
	}
}

func CreatePollDraft(dto CreatePollDraftDto) (*PollDraft, error) {
	return db.Execute(func(q *queries.Queries) (*PollDraft, error) {
		draft, err := q.CreatePollDraft(ctx.Background(), queries.CreatePollDraftParams{
			ChatID:          dto.ChatID,
			UserID:          dto.UserID,
			Command:         dto.Command,
			Title:           dto.Title,
			Flags:           FlagsToStrings(dto.PendingFlags),
			PinEnabled:      dto.PinEnabled,
			CronExpr:        pgtype.Text{String: dto.CronExpr, Valid: dto.CronExpr != ""},
			SourceMessageID: dto.SourceMessageID,
			PromptMessageID: dto.PromptMessageID,
		})
		if err != nil {
			return nil, err
		}

		return mapPollDraft(&draft), nil
	})
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

func UpdatePollDraft(dto UpdatePollDraftDto) (*PollDraft, error) {
	return db.Execute(func(q *queries.Queries) (*PollDraft, error) {
		draft, err := q.UpdatePollDraft(ctx.Background(), queries.UpdatePollDraftParams{
			ID:              dto.ID,
			Flags:           FlagsToStrings(dto.PendingFlags),
			PinEnabled:      dto.PinEnabled,
			CronExpr:        pgtype.Text{String: dto.CronExpr, Valid: dto.CronExpr != ""},
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

func TryHandleBinPollDraftReply(bot *tgbot.BotAPI, update *tgbot.Update) (bool, error) {
	message := update.Message
	if message == nil || message.ReplyToMessage == nil || message.From == nil {
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

	if len(draft.PendingFlags) == 0 {
		if err := DeletePollDraftByID(draft.ID); err != nil {
			return true, err
		}
		return true, errors2.New("Poll draft is already completed")
	}

	step := draft.PendingFlags[0]
	remaining := draft.PendingFlags[1:]
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

	if len(remaining) > 0 {
		nextPrompt, err := SendDraftPrompt(bot, message.Chat.ID, message.MessageID, DraftStepPrompt(remaining[0]))
		if err != nil {
			return true, err
		}

		_, err = UpdatePollDraft(UpdatePollDraftDto{
			ID:              draft.ID,
			PendingFlags:    remaining,
			PinEnabled:      pinEnabled,
			CronExpr:        cronExpr,
			PromptMessageID: int64(nextPrompt.MessageID),
		})
		if err != nil {
			_ = DeletePollDraftByID(draft.ID)
			return true, err
		}

		return true, nil
	}

	draft.PendingFlags = remaining
	draft.PinEnabled = pinEnabled
	draft.CronExpr = cronExpr

	finalPoll := BuildFinalPoll(draft)
	if draft.CronExpr == "" {
		if err := SendPoll(bot, &finalPoll, message); err != nil {
			return true, err
		}
		if err := DeletePollDraftByID(draft.ID); err != nil {
			return true, err
		}
		return true, nil
	}

	if err := CheckPoll(&finalPoll); err != nil {
		return true, err
	}

	storedPoll, err := CreateStoredPoll(CreateStoredPollDto{
		ChatID:   draft.ChatID,
		UserID:   user.ID,
		Command:  draft.Command,
		Poll:     finalPoll,
		CronExpr: draft.CronExpr,
	})
	if err != nil {
		return true, err
	}

	if err := RegisterCronPoll(storedPoll); err != nil {
		return true, err
	}

	if err := DeletePollDraftByID(draft.ID); err != nil {
		return true, err
	}

	_, err = SendDraftPrompt(bot, message.Chat.ID, message.MessageID, fmt2.Sprintf("Cron poll #%d was created", storedPoll.ID))
	if err != nil {
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
		PendingFlags:    StringsToFlags(dbDraft.Flags),
		PinEnabled:      dbDraft.PinEnabled,
		CronExpr:        cronExpr,
		SourceMessageID: dbDraft.SourceMessageID,
		PromptMessageID: dbDraft.PromptMessageID,
		CreatedAt:       dbDraft.CreatedAt.Time,
		UpdatedAt:       dbDraft.UpdatedAt.Time,
	}
}

func (draft *PollDraft) IsExpired() bool {
	return time2.Since(draft.CreatedAt) > draftTTL
}
