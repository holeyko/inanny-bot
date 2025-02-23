-- liquibase formatted sql

-- changeset holeko:create_flows
CREATE TABLE IF NOT EXISTS flows(
    tg_id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    context JSONB NOT NULL
);