-- name: CreatePoll :one
INSERT INTO polls (chat_id, user_id, command, title, options, flags, cron_expr)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, chat_id, user_id, command, title, options, flags, cron_expr, created_at, updated_at;

-- name: GetCronPolls :many
SELECT id, chat_id, user_id, command, title, options, flags, cron_expr, created_at, updated_at
FROM polls
WHERE cron_expr IS NOT NULL
ORDER BY id;

-- name: GetCronPollsByChatAndUser :many
SELECT id, chat_id, user_id, command, title, options, flags, cron_expr, created_at, updated_at
FROM polls
WHERE chat_id = $1 AND user_id = $2 AND cron_expr IS NOT NULL
ORDER BY id;

-- name: DeleteCronPollByIDChatAndUser :execrows
DELETE FROM polls
WHERE id = $1 AND chat_id = $2 AND user_id = $3 AND cron_expr IS NOT NULL;
