package aiquota

import (
	"context"
	"errors"
	"time"

	"github.com/holeyko/innany-tgbot/internal/generated/queries"
	"github.com/holeyko/innany-tgbot/internal/inanny/features/ai"
	"github.com/holeyko/innany-tgbot/internal/inanny/infra/db"
	"github.com/jackc/pgx/v5"
)

const DailyLimit = 20

type Store struct{}

// Reserve atomically counts dispatched attempts, including provider failures.
// The database clock defines the UTC day across all bot instances.
func (Store) Reserve(ctx context.Context, chatID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return db.ExecuteContext(ctx, func(q *queries.Queries) (bool, error) {
		_, err := q.ReserveAIRequest(ctx, chatID)
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return err == nil, err
	})
}

func (Store) Usage(ctx context.Context, chatID int64) (ai.Quota, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return db.ExecuteContext(ctx, func(q *queries.Queries) (ai.Quota, error) {
		row, err := q.GetAIQuota(ctx, chatID)
		if err != nil {
			return ai.Quota{}, err
		}
		return ai.Quota{Used: int(row.RequestsUsed), Limit: DailyLimit, ResetsAt: row.ResetsAt.Time}, nil
	})
}
