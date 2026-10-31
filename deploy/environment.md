# Deployment environment

Reference examples only. Export these variables in your shell or configure them
in CI. Deployment does not read this file or any env file. Never track credentials.

```sh
# Absent/empty URLs bootstrap local persistent stacks. No fake credentials.
unset POSTGRES_URL GRAFANA_URL
export DB_HOST=inanny-postgres
export DB_USER=inanny
export DB_PASSWORD=REPLACE_WITH_STRONG_PASSWORD
export DB_NAME=inanny
# POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB are accepted as fallbacks.
export TELEGRAM_BOT_TOKEN=REPLACE_TOKEN
export IMAGE_VERSION=latest
# Existing databases must keep their current PostgreSQL major version.
export POSTGRES_IMAGE=postgres:16
export LIQUIBASE_IMAGE=liquibase/liquibase:4.31
export GRAFANA_ADMIN=admin
export GRAFANA_PASSWORD=replace-with-a-different-strong-password
# GRAFANA_ADMIN_USER, GRAFANA_ADMIN_PASSWORD are accepted as fallbacks.
# Optional bot settings: OPENROUTER_API_KEY, OPENROUTER_MODEL, OPENAI_API_KEY,
# CHATGPT_API_TOKEN, DEEPSEEK_API_TOKEN, GOOGLE_AI_STUDIO_API_KEY, GEMINI_API_KEY,
# GOOGLE_API_KEY, AI_MAX_INPUT_CHARS, AI_MAX_OUTPUT_TOKENS,
# AI_MAX_CONCURRENT_SUMMARIES, DEBUG.
# Optional loopback ports: LOKI_PORT=3100, ALLOY_PORT=12345.
# GRAFANA_PORT=3000 controls the local Grafana loopback-only published port.
# Explicit URLs NEVER bootstrap, even localhost. URL-encode credentials.
# POSTGRES_URL=postgresql://user:password@db.example:5433/inanny?sslmode=verify-full
# localhost inside containers is not the Docker host; use host.docker.internal
# explicitly for host services (and a matching DNS name for verified TLS).
# GRAFANA_URL=https://grafana.example (entire Grafana/Loki/Alloy stack remote).
# GRAFANA_HEALTH_TOKEN= (optional bearer token for /api/health behind an auth proxy)
# DB_TLS_DIR=/absolute/path/to/certs (mounted read-only at /run/db-tls)
# MIGRATION_DB_URL=jdbc:postgresql://db.example:5433/inanny?sslmode=verify-full
# MIGRATION_DB_URL is REQUIRED for every external POSTGRES_URL. No URL translation.
# LIQUIBASE_COMMAND_USERNAME=user
# LIQUIBASE_COMMAND_PASSWORD=REPLACE_PASSWORD
# Credentials fall back to DB_USER/DB_PASSWORD, then POSTGRES_USER/POSTGRES_PASSWORD.
# Use JDBC-compatible options/certificates preserving POSTGRES_URL's security policy.
# ALLOY_DOCKER_SOCKET=/run/inanny/bot-docker.sock (logging host only; see docs)
```
