package notifications

import (
	ctx "context"
	"time"

	"github.com/holeyko/innany-tgbot/internal/generated/queries"
	"github.com/holeyko/innany-tgbot/internal/inanny/infra/db"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetNotificationByID(id int64) (*Notification, error) {
	return db.Execute(func(q *queries.Queries) (*Notification, error) {
		notification, err := q.GetNotificationByID(ctx.Background(), id)
		if err != nil {
			return nil, err
		}
		return mapNotification(&notification), nil
	})
}

func GetNotificationsByChatID(chatID int64) ([]*Notification, error) {
	return db.Execute(func(q *queries.Queries) ([]*Notification, error) {
		dbNotifications, err := q.GetNotificationsByChatID(ctx.Background(), chatID)
		if err != nil {
			return nil, err
		}

		notifications := make([]*Notification, len(dbNotifications))
		for i, notification := range dbNotifications {
			notifications[i] = mapNotification(&notification)
		}

		return notifications, nil
	})
}

type CreateNotificationDto struct {
	ChatID   int64
	Title    string
	Interval time.Duration
	EndAt    time.Time
}

func CreateNotification(dto CreateNotificationDto) (*Notification, error) {
	return db.Execute(func(q *queries.Queries) (*Notification, error) {
		interval := durationToInterval(dto.Interval)

		endAt := pgtype.Timestamp{
			Time:  dto.EndAt,
			Valid: true,
		}

		notification, err := q.CreateNotification(ctx.Background(), queries.CreateNotificationParams{
			ChatID:   dto.ChatID,
			Title:    dto.Title,
			Interval: interval,
			EndAt:    endAt,
		})

		if err != nil {
			return nil, err
		}

		return mapNotification(&notification), nil
	})
}

func DeleteNotification(id int64) error {
	_, err := db.Execute(func(q *queries.Queries) (struct{}, error) {
		return struct{}{}, q.DeleteNotification(ctx.Background(), id)
	})
	return err
}
