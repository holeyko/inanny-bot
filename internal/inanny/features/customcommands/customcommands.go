package customcommand

import (
	ctx "context"
	"errors"
	"fmt"
	log2 "log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/holeyko/innany-tgbot/internal/generated/queries"
	"github.com/holeyko/innany-tgbot/internal/inanny/infra/db"
	"github.com/jackc/pgx/v5"
)

const draftTTL = time.Hour

var (
	ErrCustomCommandNotFound      = errors.New("custom command not found")
	ErrCustomCommandDraftNotFound = errors.New("custom command draft not found")
	paramRegexp                   = regexp.MustCompile(`\$(\d+)`)
)

type UserDto struct {
	TelegramLogin string
	FirstName     string
	LastName      string
}

type CustomCommand struct {
	ID            int64
	ChatID        int64
	UserID        int64
	Name          string
	TargetCommand string
	Body          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CustomCommandDraft struct {
	ID              int64
	ChatID          int64
	UserID          int64
	PromptMessageID int64
	CreatedAt       time.Time
}

type CreateCustomCommandDto struct {
	ChatID        int64
	UserID        int64
	Name          string
	TargetCommand string
	Body          string
}

type CommandTemplate struct {
	Name          string
	TargetCommand string
	Body          string
}

func UpsertUser(dto UserDto) (*queries.User, error) {
	if dto.TelegramLogin == "" {
		return nil, errors.New("Telegram username is required to manage custom commands")
	}

	return db.Execute(func(q *queries.Queries) (*queries.User, error) {
		user, err := q.UpsertUserByTelegramLogin(ctx.Background(), queries.UpsertUserByTelegramLoginParams{
			TelegramLogin: dto.TelegramLogin,
			FirstName:     dto.FirstName,
			LastName:      dto.LastName,
		})
		if err != nil {
			return nil, err
		}
		return &user, nil
	})
}

func ParseCommandTemplate(input string) (CommandTemplate, error) {
	input = strings.TrimSpace(input)
	name, remain, ok := cutToken(input)
	if !ok {
		return CommandTemplate{}, errors.New("Usage: <command_name> <existing_command> <command_body>")
	}

	targetCommand, body, ok := cutToken(remain)
	if !ok {
		return CommandTemplate{}, errors.New("Usage: <command_name> <existing_command> <command_body>")
	}

	name = strings.TrimPrefix(name, "/")
	targetCommand = strings.TrimPrefix(targetCommand, "/")
	body = strings.TrimSpace(body)
	if body == "" {
		return CommandTemplate{}, errors.New("Command body can't be empty")
	}
	if !isValidCommandName(name) {
		return CommandTemplate{}, fmt.Errorf("Invalid command name: %s", name)
	}
	if !isValidCommandName(targetCommand) {
		return CommandTemplate{}, fmt.Errorf("Invalid existing command: %s", targetCommand)
	}

	return CommandTemplate{Name: name, TargetCommand: targetCommand, Body: body}, nil
}

func MaxRequiredParameter(body string) int {
	matches := paramRegexp.FindAllStringSubmatch(body, -1)
	maxParam := 0
	for _, match := range matches {
		value, err := strconv.Atoi(match[1])
		if err == nil && value > maxParam {
			maxParam = value
		}
	}
	return maxParam
}

func ApplyParameters(command *CustomCommand, args []string) (string, error) {
	required := MaxRequiredParameter(command.Body)
	if len(args) < required {
		return "", fmt.Errorf(
			"Command /%s expects at least %d parameter(s), got %d. Usage: /%s %s",
			command.Name,
			required,
			len(args),
			command.Name,
			usageParams(required),
		)
	}

	body := paramRegexp.ReplaceAllStringFunc(command.Body, func(match string) string {
		index, err := strconv.Atoi(strings.TrimPrefix(match, "$"))
		if err != nil || index <= 0 || index > len(args) {
			return match
		}
		return args[index-1]
	})

	return body, nil
}

func CreateCustomCommand(dto CreateCustomCommandDto) (*CustomCommand, error) {
	return db.Execute(func(q *queries.Queries) (*CustomCommand, error) {
		command, err := q.CreateCustomCommand(ctx.Background(), queries.CreateCustomCommandParams{
			ChatID:        dto.ChatID,
			UserID:        dto.UserID,
			Name:          dto.Name,
			TargetCommand: dto.TargetCommand,
			Body:          dto.Body,
		})
		if err != nil {
			return nil, err
		}
		return mapCustomCommand(&command), nil
	})
}

func GetCustomCommandByChatAndName(chatID int64, name string) (*CustomCommand, error) {
	return db.Execute(func(q *queries.Queries) (*CustomCommand, error) {
		command, err := q.GetCustomCommandByChatAndName(ctx.Background(), queries.GetCustomCommandByChatAndNameParams{
			ChatID: chatID,
			Name:   name,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrCustomCommandNotFound
			}
			return nil, err
		}
		return mapCustomCommand(&command), nil
	})
}

func ListCustomCommandsByChat(chatID int64) ([]*CustomCommand, error) {
	return db.Execute(func(q *queries.Queries) ([]*CustomCommand, error) {
		dbCommands, err := q.ListCustomCommandsByChat(ctx.Background(), chatID)
		if err != nil {
			return nil, err
		}
		commands := make([]*CustomCommand, len(dbCommands))
		for i, command := range dbCommands {
			commands[i] = mapCustomCommand(&command)
		}
		return commands, nil
	})
}

func DeleteCustomCommandByIDChatAndUser(id int64, chatID int64, userID int64) (bool, error) {
	rowsAffected, err := db.Execute(func(q *queries.Queries) (int64, error) {
		return q.DeleteCustomCommandByIDChatAndUser(ctx.Background(), queries.DeleteCustomCommandByIDChatAndUserParams{
			ID:     id,
			ChatID: chatID,
			UserID: userID,
		})
	})
	return rowsAffected > 0, err
}

func DeleteCustomCommandByIDAndChat(id int64, chatID int64) (bool, error) {
	rowsAffected, err := db.Execute(func(q *queries.Queries) (int64, error) {
		return q.DeleteCustomCommandByIDAndChat(ctx.Background(), queries.DeleteCustomCommandByIDAndChatParams{
			ID:     id,
			ChatID: chatID,
		})
	})
	return rowsAffected > 0, err
}

func CreateCustomCommandDraft(chatID int64, userID int64, promptMessageID int64) (*CustomCommandDraft, error) {
	return db.Execute(func(q *queries.Queries) (*CustomCommandDraft, error) {
		draft, err := q.CreateCustomCommandDraft(ctx.Background(), queries.CreateCustomCommandDraftParams{
			ChatID:          chatID,
			UserID:          userID,
			PromptMessageID: promptMessageID,
		})
		if err != nil {
			return nil, err
		}
		return mapCustomCommandDraft(&draft), nil
	})
}

func GetCustomCommandDraftByPromptMessageID(chatID int64, userID int64, promptMessageID int64) (*CustomCommandDraft, error) {
	return db.Execute(func(q *queries.Queries) (*CustomCommandDraft, error) {
		draft, err := q.GetCustomCommandDraftByPromptMessageID(ctx.Background(), queries.GetCustomCommandDraftByPromptMessageIDParams{
			ChatID:          chatID,
			UserID:          userID,
			PromptMessageID: promptMessageID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrCustomCommandDraftNotFound
			}
			return nil, err
		}
		return mapCustomCommandDraft(&draft), nil
	})
}

func DeleteCustomCommandDraftByID(id int64) error {
	_, err := db.Execute(func(q *queries.Queries) (struct{}, error) {
		_, err := q.DeleteCustomCommandDraftByID(ctx.Background(), id)
		return struct{}{}, err
	})
	return err
}

func CleanupExpiredCustomCommandDrafts() (int64, error) {
	return db.Execute(func(q *queries.Queries) (int64, error) {
		return q.DeleteExpiredCustomCommandDrafts(ctx.Background())
	})
}

func StartDraftCleanup() {
	if _, err := CleanupExpiredCustomCommandDrafts(); err != nil {
		log2.Println("Error while cleaning expired custom command drafts:", err)
	}

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if _, err := CleanupExpiredCustomCommandDrafts(); err != nil {
				log2.Println("Error while cleaning expired custom command drafts:", err)
			}
		}
	}()
}

func (draft *CustomCommandDraft) IsExpired() bool {
	return time.Since(draft.CreatedAt) > draftTTL
}

func cutToken(input string) (string, string, bool) {
	input = strings.TrimLeftFunc(input, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' })
	if input == "" {
		return "", "", false
	}

	for i, r := range input {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return input[:i], input[i+1:], true
		}
	}

	return "", "", false
}

func isValidCommandName(command string) bool {
	if command == "" || len(command) > 64 {
		return false
	}
	for _, r := range command {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}

func usageParams(count int) string {
	params := make([]string, count)
	for i := range params {
		params[i] = fmt.Sprintf("<%d>", i+1)
	}
	return strings.Join(params, " ")
}

func mapCustomCommand(dbCommand *queries.CustomCommand) *CustomCommand {
	return &CustomCommand{
		ID:            dbCommand.ID,
		ChatID:        dbCommand.ChatID,
		UserID:        dbCommand.UserID,
		Name:          dbCommand.Name,
		TargetCommand: dbCommand.TargetCommand,
		Body:          dbCommand.Body,
		CreatedAt:     dbCommand.CreatedAt.Time,
		UpdatedAt:     dbCommand.UpdatedAt.Time,
	}
}

func mapCustomCommandDraft(dbDraft *queries.CustomCommandDraft) *CustomCommandDraft {
	return &CustomCommandDraft{
		ID:              dbDraft.ID,
		ChatID:          dbDraft.ChatID,
		UserID:          dbDraft.UserID,
		PromptMessageID: dbDraft.PromptMessageID,
		CreatedAt:       dbDraft.CreatedAt.Time,
	}
}
