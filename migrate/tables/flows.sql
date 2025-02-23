-- liquibase formatted sql

-- changeset holeko:create_flows
CREATE TABLE IF NOT EXISTS flows(
    tg_id BIGINT NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    context JSONB NOT NULL,

    PRIMARY KEY (tg_id, "name")
);