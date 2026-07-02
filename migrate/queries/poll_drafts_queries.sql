-- name: CreatePollDraft :one
INSERT INTO poll_drafts (chat_id, user_id, command, title, flags, pin_enabled, cron_expr, source_message_id, prompt_message_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, chat_id, user_id, command, title, flags, pin_enabled, cron_expr, source_message_id, prompt_message_id, created_at, updated_at;

-- name: GetPollDraftByPromptMessageID :one
SELECT id, chat_id, user_id, command, title, flags, pin_enabled, cron_expr, source_message_id, prompt_message_id, created_at, updated_at
FROM poll_drafts
WHERE chat_id = $1 AND user_id = $2 AND prompt_message_id = $3;

-- name: UpdatePollDraft :one
UPDATE poll_drafts
SET flags = $2,
    pin_enabled = $3,
    cron_expr = $4,
    prompt_message_id = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING id, chat_id, user_id, command, title, flags, pin_enabled, cron_expr, source_message_id, prompt_message_id, created_at, updated_at;

-- name: DeletePollDraftByID :execrows
DELETE FROM poll_drafts
WHERE id = $1;

-- name: DeleteExpiredPollDrafts :execrows
DELETE FROM poll_drafts
WHERE created_at < NOW() - INTERVAL '1 hour';
