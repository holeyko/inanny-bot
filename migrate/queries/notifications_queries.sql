-- name: GetNotificationByID :one
SELECT * FROM notifications WHERE id = $1;

-- name: GetNotificationsByChatID :many
SELECT * FROM notifications WHERE chat_id = $1;

-- name: CreateNotification :one
INSERT INTO notifications (chat_id, title, interval, end_at) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeleteNotification :exec
DELETE FROM notifications WHERE id = $1;
