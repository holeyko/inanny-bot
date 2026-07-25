--liquibase formatted sql

--changeset holeyko:add_messages_table runInTransaction:false runOnChange:false
CREATE TABLE IF NOT EXISTS messages (
    chat_id BIGINT NOT NULL,
    message_id BIGINT NOT NULL,
    sender VARCHAR(256) NOT NULL DEFAULT '',
    text TEXT NOT NULL,
    reply_to_message_id BIGINT,
    sent_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chat_id, message_id)
);

CREATE INDEX IF NOT EXISTS idx_messages_chat_message ON messages (chat_id, message_id);
