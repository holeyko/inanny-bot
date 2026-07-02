-- name: UpsertUserByTelegramLogin :one
INSERT INTO users (telegram_login, first_name, last_name)
VALUES ($1, $2, $3)
ON CONFLICT (telegram_login) DO UPDATE SET
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name
RETURNING id, telegram_login, first_name, last_name, created_at;
