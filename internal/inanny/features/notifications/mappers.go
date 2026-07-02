package notifications

import "github.com/holeyko/innany-tgbot/internal/generated/queries"

func mapNotification(notification *queries.Notification) *Notification {
	return &Notification{
		ID:       notification.ID,
		ChatID:   notification.ChatID,
		Title:    notification.Title,
		Interval: intervalToDuration(notification.Interval),
		EndAt:    notification.EndAt.Time,
	}
}
