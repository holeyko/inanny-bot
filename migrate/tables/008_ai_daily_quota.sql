--liquibase formatted sql

--changeset holeyko:add_ai_daily_quota runInTransaction:false runOnChange:false
CREATE TABLE IF NOT EXISTS ai_daily_quota (
    chat_id BIGINT NOT NULL,
    usage_date DATE NOT NULL,
    requests_used INTEGER NOT NULL CHECK (requests_used BETWEEN 1 AND 20),
    PRIMARY KEY (chat_id, usage_date)
);
