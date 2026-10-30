-- name: ReserveAIRequest :one
INSERT INTO ai_daily_quota (chat_id, usage_date, requests_used)
VALUES ($1, (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date, 1)
ON CONFLICT (chat_id, usage_date) DO UPDATE
SET requests_used = ai_daily_quota.requests_used + 1
WHERE ai_daily_quota.requests_used < 20
RETURNING requests_used;

-- name: GetAIQuota :one
SELECT COALESCE((
    SELECT requests_used FROM ai_daily_quota
    WHERE chat_id = $1 AND usage_date = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date
), 0)::integer AS requests_used,
(((CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date + 1)::timestamp AT TIME ZONE 'UTC')::timestamptz AS resets_at;
