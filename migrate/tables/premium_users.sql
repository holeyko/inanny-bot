-- liquibase formatted sql


-- changeset holeyko:create_premium_users
CREATE TABLE IF NOT EXISTS premium_users(
    tg_id BIGINT PRIMARY KEY
);