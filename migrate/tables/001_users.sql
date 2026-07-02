--liquibase formatted sql

--changeset holeyko:add_users_table runInTransaction:false runOnChange:false
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    telegram_login VARCHAR(256) NOT NULL UNIQUE,
    first_name VARCHAR(256) NOT NULL DEFAULT '',
    last_name VARCHAR(256) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
