package messages

import (
	"context"
	"errors"
	"time"

	"github.com/holeyko/innany-tgbot/internal/generated/queries"
	"github.com/holeyko/innany-tgbot/internal/inanny/infra/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type Message struct {
	ChatID           int64
	MessageID        int64
	Sender           string
	Text             string
	ReplyToMessageID int64
	SentAt           time.Time
}

type Repository interface {
	Save(message Message) error
	ListBetween(chatID int64, fromMessageID int64, beforeMessageID int64) ([]Message, error)
}

type repository struct{}

func NewRepository() Repository {
	return repository{}
}

func (repository) Save(message Message) error {
	if message.ChatID == 0 || message.MessageID == 0 || message.Text == "" {
		return errors.New("message must have chat id, message id, and text")
	}

	_, err := db.Execute(func(q *queries.Queries) (struct{}, error) {
		err := q.UpsertMessage(context.Background(), queries.UpsertMessageParams{
			ChatID:           message.ChatID,
			MessageID:        message.MessageID,
			Sender:           message.Sender,
			Text:             message.Text,
			ReplyToMessageID: pgtype.Int8{Int64: message.ReplyToMessageID, Valid: message.ReplyToMessageID != 0},
			SentAt:           pgtype.Timestamp{Time: message.SentAt, Valid: !message.SentAt.IsZero()},
		})
		return struct{}{}, err
	})
	return err
}

func (repository) ListBetween(chatID int64, fromMessageID int64, beforeMessageID int64) ([]Message, error) {
	return db.Execute(func(q *queries.Queries) ([]Message, error) {
		storedMessages, err := q.ListMessagesBetween(context.Background(), queries.ListMessagesBetweenParams{
			ChatID:      chatID,
			MessageID:   fromMessageID,
			MessageID_2: beforeMessageID,
		})
		if err != nil {
			return nil, err
		}

		result := make([]Message, len(storedMessages))
		for i, storedMessage := range storedMessages {
			result[i] = Message{
				ChatID:           storedMessage.ChatID,
				MessageID:        storedMessage.MessageID,
				Sender:           storedMessage.Sender,
				Text:             storedMessage.Text,
				ReplyToMessageID: int8Value(storedMessage.ReplyToMessageID),
				SentAt:           storedMessage.SentAt.Time,
			}
		}
		return result, nil
	})
}

func int8Value(value pgtype.Int8) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}
