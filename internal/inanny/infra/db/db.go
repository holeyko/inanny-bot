package db

import (
	"context"
	"fmt"
	"os"

	"github.com/holeyko/innany-tgbot/internal/generated/queries"
	"github.com/jackc/pgx/v5"
)

var (
	host     = os.Getenv("DB_HOST")
	user     = os.Getenv("DB_USER")
	password = os.Getenv("DB_PASSWORD")
	dbname   = os.Getenv("DB_NAME")

	connectionString = fmt.Sprintf("postgres://%s:%s@%s/%s", user, password, host, dbname)
)

func Execute[T any](f func(q *queries.Queries) (T, error)) (T, error) {
	var result T

	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		return result, fmt.Errorf("exception while pooling connection: %w", err)
	}
	defer conn.Close(context.Background())

	queries := queries.New(conn)
	result, err = f(queries)

	return result, err
}

func ExecuteInTransaction[T any](f func(q *queries.Queries) (T, error)) (T, error) {
	var result T

	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		return result, fmt.Errorf("exception while pooling connection: %w", err)
	}
	defer conn.Close(context.Background())

	tx, err := conn.Begin(context.Background())
	if err != nil {
		return result, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	queries := queries.New(tx)
	result, err = f(queries)
	if err != nil {
		return result, err
	}

	err = tx.Commit(context.Background())
	return result, err
}
