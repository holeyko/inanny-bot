--liquibase formatted sql

--changeset holeyko:add_polls_table runInTransaction:false runOnChange:false
CREATE TABLE IF NOT EXISTS polls (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id),
    command VARCHAR(64) NOT NULL,
    title VARCHAR(512) NOT NULL,
    options TEXT[] NOT NULL,
    flags TEXT[] NOT NULL DEFAULT '{}',
    cron_expr VARCHAR(128),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_polls_chat_user ON polls (chat_id, user_id);
CREATE INDEX IF NOT EXISTS idx_polls_cron ON polls (cron_expr) WHERE cron_expr IS NOT NULL;
