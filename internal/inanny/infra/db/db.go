package db

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/holeyko/innany-tgbot/internal/generated/queries"
	"github.com/jackc/pgx/v5"
)

var connectionString = databaseURL()

func databaseURL() string {
	if explicit := os.Getenv("POSTGRES_URL"); explicit != "" {
		return explicit // pgx handles URI credentials, ports and options natively.
	}
	connection := url.URL{
		Scheme: "postgres",
		Host:   os.Getenv("DB_HOST"),
		User:   url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")),
		Path:   "/" + os.Getenv("DB_NAME"),
	}
	return connection.String()
}

// Ping verifies connectivity using the same configuration as application queries.
func Ping(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	return conn.Ping(ctx)
}

func Execute[T any](f func(q *queries.Queries) (T, error)) (T, error) {
	return ExecuteContext(context.Background(), f)
}

// ExecuteContext bounds both connection establishment and the caller's queries.
func ExecuteContext[T any](ctx context.Context, f func(q *queries.Queries) (T, error)) (T, error) {
	var result T

	conn, err := pgx.Connect(ctx, connectionString)
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
