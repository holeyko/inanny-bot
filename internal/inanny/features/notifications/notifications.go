package notifications

import "time"

type Notification struct {
	ID       int64
	ChatID   int64
	Title    string
	Interval time.Duration
	EndAt    time.Time
}
