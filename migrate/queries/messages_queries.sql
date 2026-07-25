-- name: UpsertMessage :exec
INSERT INTO messages (
    chat_id,
    message_id,
    sender,
    text,
    reply_to_message_id,
    sent_at
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (chat_id, message_id) DO UPDATE SET
    sender = EXCLUDED.sender,
    text = EXCLUDED.text,
    reply_to_message_id = EXCLUDED.reply_to_message_id,
    sent_at = EXCLUDED.sent_at;

-- name: ListMessagesBetween :many
SELECT chat_id, message_id, sender, text, reply_to_message_id, sent_at, created_at
FROM messages
WHERE chat_id = $1
  AND message_id >= $2
  AND message_id < $3
ORDER BY message_id;
