--liquibase formatted sql

--changeset holeyko:add_custom_commands_table runInTransaction:false runOnChange:false
CREATE TABLE IF NOT EXISTS custom_commands (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id),
    name VARCHAR(64) NOT NULL,
    target_command VARCHAR(64) NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT custom_commands_chat_name_unique UNIQUE (chat_id, name)
);

CREATE INDEX IF NOT EXISTS idx_custom_commands_chat ON custom_commands (chat_id);

--changeset holeyko:add_custom_command_drafts_table runInTransaction:false runOnChange:false
CREATE TABLE IF NOT EXISTS custom_command_drafts (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id),
    prompt_message_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_custom_command_drafts_prompt ON custom_command_drafts (chat_id, user_id, prompt_message_id);
CREATE INDEX IF NOT EXISTS idx_custom_command_drafts_created_at ON custom_command_drafts (created_at);
