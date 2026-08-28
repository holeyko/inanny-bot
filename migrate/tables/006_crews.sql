--liquibase formatted sql

--changeset holeyko:add_crews_tables runInTransaction:false runOnChange:false
CREATE TABLE IF NOT EXISTS crews (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    chat_id BIGINT NOT NULL,
    creator_user_id BIGINT NOT NULL REFERENCES users(id),
    name VARCHAR(32) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT crews_chat_name_unique UNIQUE (chat_id, name)
);

CREATE TABLE IF NOT EXISTS crew_members (
    crew_id BIGINT NOT NULL REFERENCES crews(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    PRIMARY KEY (crew_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_crews_chat ON crews (chat_id);
CREATE INDEX IF NOT EXISTS idx_crew_members_user ON crew_members (user_id);
