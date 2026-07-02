package poll

import (
	ctx "context"
	"errors"

	"github.com/holeyko/innany-tgbot/internal/generated/queries"
	"github.com/holeyko/innany-tgbot/internal/inanny/infra/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserDto struct {
	TelegramLogin string
	FirstName     string
	LastName      string
}

func UpsertUser(dto UserDto) (*queries.User, error) {
	if dto.TelegramLogin == "" {
		return nil, errors.New("Telegram username is required to manage cron polls")
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

type CreateStoredPollDto struct {
	ChatID   int64
	UserID   int64
	Command  string
	Poll     Poll
	CronExpr string
}

func CreateStoredPoll(dto CreateStoredPollDto) (*StoredPoll, error) {
	return db.Execute(func(q *queries.Queries) (*StoredPoll, error) {
		poll, err := q.CreatePoll(ctx.Background(), queries.CreatePollParams{
			ChatID:   dto.ChatID,
			UserID:   dto.UserID,
			Command:  dto.Command,
			Title:    dto.Poll.Title,
			Options:  dto.Poll.Options,
			Flags:    FlagsToStrings(dto.Poll.Flags),
			CronExpr: pgtype.Text{String: dto.CronExpr, Valid: dto.CronExpr != ""},
		})
		if err != nil {
			return nil, err
		}
		return mapStoredPoll(&poll), nil
	})
}

func GetCronPolls() ([]*StoredPoll, error) {
	return db.Execute(func(q *queries.Queries) ([]*StoredPoll, error) {
		dbPolls, err := q.GetCronPolls(ctx.Background())
		if err != nil {
			return nil, err
		}

		polls := make([]*StoredPoll, len(dbPolls))
		for i, dbPoll := range dbPolls {
			polls[i] = mapStoredPoll(&dbPoll)
		}
		return polls, nil
	})
}

func GetCronPollsByChatAndUser(chatID int64, userID int64) ([]*StoredPoll, error) {
	return db.Execute(func(q *queries.Queries) ([]*StoredPoll, error) {
		dbPolls, err := q.GetCronPollsByChatAndUser(ctx.Background(), queries.GetCronPollsByChatAndUserParams{
			ChatID: chatID,
			UserID: userID,
		})
		if err != nil {
			return nil, err
		}

		polls := make([]*StoredPoll, len(dbPolls))
		for i, dbPoll := range dbPolls {
			polls[i] = mapStoredPoll(&dbPoll)
		}
		return polls, nil
	})
}

func DeleteCronPollByIDChatAndUser(id int64, chatID int64, userID int64) (bool, error) {
	rowsAffected, err := db.Execute(func(q *queries.Queries) (int64, error) {
		return q.DeleteCronPollByIDChatAndUser(ctx.Background(), queries.DeleteCronPollByIDChatAndUserParams{
			ID:     id,
			ChatID: chatID,
			UserID: userID,
		})
	})
	return rowsAffected > 0, err
}

func mapStoredPoll(dbPoll *queries.Poll) *StoredPoll {
	cronExpr := ""
	if dbPoll.CronExpr.Valid {
		cronExpr = dbPoll.CronExpr.String
	}

	return &StoredPoll{
		Poll: Poll{
			ID:      dbPoll.ID,
			ChatID:  dbPoll.ChatID,
			Command: dbPoll.Command,
			Title:   dbPoll.Title,
			Options: dbPoll.Options,
			Flags:   StringsToFlags(dbPoll.Flags),
		},
		CronExpr:  cronExpr,
		CreatedAt: dbPoll.CreatedAt.Time,
	}
}
