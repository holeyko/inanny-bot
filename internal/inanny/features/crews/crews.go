package crew

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/holeyko/innany-tgbot/internal/generated/queries"
	"github.com/holeyko/innany-tgbot/internal/inanny/infra/db"
	"github.com/jackc/pgx/v5"
)

var ErrCrewNotFound = errors.New("crew not found")

var telegramNameRegexp = regexp.MustCompile(`^[A-Za-z0-9_]{1,32}$`)

type UserDto struct {
	TelegramLogin string
	FirstName     string
	LastName      string
}

type Crew struct {
	ID            int64
	ChatID        int64
	CreatorUserID int64
	Name          string
	Members       []string
}

type CreateCrewDto struct {
	ChatID  int64
	Creator UserDto
	Name    string
	Members []string
}

func UpsertUser(dto UserDto) (*queries.User, error) {
	if dto.TelegramLogin == "" {
		return nil, errors.New("Telegram username is required to manage crews")
	}

	return db.Execute(func(q *queries.Queries) (*queries.User, error) {
		user, err := q.UpsertUserByTelegramLogin(context.Background(), queries.UpsertUserByTelegramLoginParams{
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

func CreateCrew(dto CreateCrewDto) (*Crew, error) {
	return db.ExecuteInTransaction(func(q *queries.Queries) (*Crew, error) {
		creator, err := q.UpsertUserByTelegramLogin(context.Background(), queries.UpsertUserByTelegramLoginParams{
			TelegramLogin: dto.Creator.TelegramLogin,
			FirstName:     dto.Creator.FirstName,
			LastName:      dto.Creator.LastName,
		})
		if err != nil {
			return nil, err
		}

		createdCrew, err := q.CreateCrew(context.Background(), queries.CreateCrewParams{
			ChatID:        dto.ChatID,
			CreatorUserID: creator.ID,
			Name:          dto.Name,
		})
		if err != nil {
			return nil, err
		}

		members := make([]string, 0, len(dto.Members)+1)
		seenMembers := make(map[string]struct{}, len(dto.Members)+1)
		for _, member := range append([]string{dto.Creator.TelegramLogin}, dto.Members...) {
			member = NormalizeUsername(member)
			if _, exists := seenMembers[member]; exists {
				continue
			}
			seenMembers[member] = struct{}{}
			members = append(members, member)
		}

		for _, member := range members {
			user, err := q.EnsureUserByTelegramLogin(context.Background(), member)
			if err != nil {
				return nil, err
			}
			if _, err := q.AddCrewMember(context.Background(), queries.AddCrewMemberParams{
				CrewID: createdCrew.ID,
				UserID: user.ID,
			}); err != nil {
				return nil, err
			}
		}

		return &Crew{
			ID:            createdCrew.ID,
			ChatID:        createdCrew.ChatID,
			CreatorUserID: createdCrew.CreatorUserID,
			Name:          createdCrew.Name,
		}, nil
	})
}

func GetCrewByChatAndName(chatID int64, name string) (*Crew, error) {
	return db.Execute(func(q *queries.Queries) (*Crew, error) {
		storedCrew, err := q.GetCrewByChatAndName(context.Background(), queries.GetCrewByChatAndNameParams{
			ChatID: chatID,
			Name:   name,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrCrewNotFound
			}
			return nil, err
		}

		members, err := q.ListCrewMembersByID(context.Background(), storedCrew.ID)
		if err != nil {
			return nil, err
		}

		return &Crew{
			ID:            storedCrew.ID,
			ChatID:        storedCrew.ChatID,
			CreatorUserID: storedCrew.CreatorUserID,
			Name:          storedCrew.Name,
			Members:       members,
		}, nil
	})
}

func AddCrewMember(crewID int64, username string) (bool, error) {
	rowsAffected, err := db.ExecuteInTransaction(func(q *queries.Queries) (int64, error) {
		user, err := q.EnsureUserByTelegramLogin(context.Background(), username)
		if err != nil {
			return 0, err
		}
		return q.AddCrewMember(context.Background(), queries.AddCrewMemberParams{
			CrewID: crewID,
			UserID: user.ID,
		})
	})
	return rowsAffected > 0, err
}

func DeleteCrewMember(crewID int64, username string) (bool, error) {
	rowsAffected, err := db.Execute(func(q *queries.Queries) (int64, error) {
		return q.DeleteCrewMember(context.Background(), queries.DeleteCrewMemberParams{
			CrewID:        crewID,
			TelegramLogin: username,
		})
	})
	return rowsAffected > 0, err
}

func DeleteCrew(crewID int64, chatID int64) (bool, error) {
	rowsAffected, err := db.Execute(func(q *queries.Queries) (int64, error) {
		return q.DeleteCrewByIDAndChat(context.Background(), queries.DeleteCrewByIDAndChatParams{
			ID:     crewID,
			ChatID: chatID,
		})
	})
	return rowsAffected > 0, err
}

func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(username), "@"))
}

func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func ValidateName(name string) error {
	if !telegramNameRegexp.MatchString(name) {
		return fmt.Errorf("Invalid crew name %q. Use 1-32 letters, numbers, or underscores", name)
	}
	return nil
}

func ValidateUsername(username string) error {
	if !telegramNameRegexp.MatchString(username) {
		return fmt.Errorf("Invalid Telegram username %q. Use 1-32 letters, numbers, or underscores", username)
	}
	return nil
}
