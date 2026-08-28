-- name: CreateCrew :one
INSERT INTO crews (chat_id, creator_user_id, name)
VALUES ($1, $2, $3)
RETURNING id, chat_id, creator_user_id, name, created_at;

-- name: GetCrewByChatAndName :one
SELECT id, chat_id, creator_user_id, name, created_at
FROM crews
WHERE chat_id = $1 AND name = $2;

-- name: ListCrewMembersByID :many
SELECT u.telegram_login
FROM crew_members cm
JOIN users u ON u.id = cm.user_id
WHERE cm.crew_id = $1
ORDER BY u.telegram_login;

-- name: AddCrewMember :execrows
INSERT INTO crew_members (crew_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteCrewMember :execrows
DELETE FROM crew_members cm
USING users u
WHERE cm.crew_id = $1
  AND cm.user_id = u.id
  AND u.telegram_login = $2;

-- name: DeleteCrewByIDAndChat :execrows
DELETE FROM crews
WHERE id = $1 AND chat_id = $2;

-- name: EnsureUserByTelegramLogin :one
INSERT INTO users (telegram_login)
VALUES ($1)
ON CONFLICT (telegram_login) DO UPDATE SET
    telegram_login = EXCLUDED.telegram_login
RETURNING id, telegram_login, first_name, last_name, created_at;
