package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	host     = os.Getenv("DB_HOST")
	user     = os.Getenv("DB_USER")
	password = os.Getenv("DB_PASSWORD")
	dbname   = os.Getenv("DB_NAME")

	connectionString = fmt.Sprintf("postgres://%s:%s@%s/%s", user, password, host, dbname)
)

func Execute[T any](f func(pool *pgxpool.Pool) T) T {
	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		log.Fatalln("Exception while pooling connection", err)
	}

	defer pool.Close()

	result := f(pool)

	return result
}

func ExecuteInTransaction[T any](f func(tx *pgx.Tx) T) T {
	return Execute(func(pool *pgxpool.Pool) T {
		tx, err := pool.Begin(context.Background())
		if err != nil {
			log.Fatalln("Can't receive transaction", err)
		}

		result := f(&tx)

		if !tx.Conn().IsClosed() {
			tx.Rollback(context.Background())
		}

		return result
	})
}
