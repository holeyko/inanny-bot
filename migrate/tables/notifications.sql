--liquibase formatted sql

--changeset holeyko:add_notifications_table runInTransaction:false runOnChange:false
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    chat_id BIGINT NOT NULL,
    title VARCHAR(256) NOT NULL,
    interval INTERVAL NOT NULL,
    end_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_notifications_chat_id ON notifications (chat_id);
