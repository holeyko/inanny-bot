--liquibase formatted sql

--changeset holeyko:add_poll_drafts_table runInTransaction:false runOnChange:false
CREATE TABLE IF NOT EXISTS poll_drafts (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id),
    command VARCHAR(64) NOT NULL,
    title VARCHAR(512) NOT NULL,
    options TEXT[] NOT NULL DEFAULT '{}',
    flags TEXT[] NOT NULL DEFAULT '{}',
    pin_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    cron_expr VARCHAR(128),
    step_index INT NOT NULL DEFAULT 0,
    source_message_id BIGINT NOT NULL,
    prompt_message_id BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_poll_drafts_chat_user ON poll_drafts (chat_id, user_id);
CREATE INDEX IF NOT EXISTS idx_poll_drafts_prompt ON poll_drafts (chat_id, user_id, prompt_message_id);
CREATE INDEX IF NOT EXISTS idx_poll_drafts_created_at ON poll_drafts (created_at);
