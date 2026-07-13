-- name: CreateCustomCommand :one
INSERT INTO custom_commands (chat_id, user_id, name, target_command, body)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, chat_id, user_id, name, target_command, body, created_at, updated_at;

-- name: GetCustomCommandByChatAndName :one
SELECT id, chat_id, user_id, name, target_command, body, created_at, updated_at
FROM custom_commands
WHERE chat_id = $1 AND name = $2;

-- name: ListCustomCommandsByChat :many
SELECT id, chat_id, user_id, name, target_command, body, created_at, updated_at
FROM custom_commands
WHERE chat_id = $1
ORDER BY name;

-- name: DeleteCustomCommandByIDChatAndUser :execrows
DELETE FROM custom_commands
WHERE id = $1 AND chat_id = $2 AND user_id = $3;

-- name: DeleteCustomCommandByIDAndChat :execrows
DELETE FROM custom_commands
WHERE id = $1 AND chat_id = $2;

-- name: CreateCustomCommandDraft :one
INSERT INTO custom_command_drafts (chat_id, user_id, prompt_message_id)
VALUES ($1, $2, $3)
RETURNING id, chat_id, user_id, prompt_message_id, created_at;

-- name: GetCustomCommandDraftByPromptMessageID :one
SELECT id, chat_id, user_id, prompt_message_id, created_at
FROM custom_command_drafts
WHERE chat_id = $1 AND user_id = $2 AND prompt_message_id = $3;

-- name: DeleteCustomCommandDraftByID :execrows
DELETE FROM custom_command_drafts
WHERE id = $1;

-- name: DeleteExpiredCustomCommandDrafts :execrows
DELETE FROM custom_command_drafts
WHERE created_at < NOW() - INTERVAL '1 hour';
