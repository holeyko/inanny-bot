MAIN_PATH = ./cmd/inanny/main.go
BUILD_DIR = ./build
APP = app
BOT_STARTUP_TIMEOUT ?= 60

.PHONY: clean build run bar
clean:
	go clean

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

bar: build run

# All orchestration is here; URLs are consumed by the native clients.
COMPOSE = docker compose
override COMPOSE_DISABLE_ENV_FILE := true
export COMPOSE_DISABLE_ENV_FILE
unexport COMPOSE_ENV_FILES
BOT = $(COMPOSE) -p inanny -f deploy/bot/docker-compose.yaml
DB = $(COMPOSE) -p inanny -f deploy/postgresql/docker-compose.yaml
GRAFANA = $(COMPOSE) -p inanny-observability -f deploy/grafana/docker-compose.yaml
IMAGE = $(COMPOSE) -p inanny -f deploy/compose.build.yaml
export BOT_STARTUP_TIMEOUT
# These projects intentionally share ownership; never remove the other stack.
override COMPOSE_IGNORE_ORPHANS := true
override COMPOSE_REMOVE_ORPHANS := false
export COMPOSE_IGNORE_ORPHANS COMPOSE_REMOVE_ORPHANS

.PHONY: deploy docker-deploy docker-bar docker-run compose-check docker-build \
	db-bootstrap db-migrate logging-up logging-status deployment-check docker-network \
	docker-ready docker-diagnostics docker-clean
deploy docker-deploy docker-bar docker-run:
	@$(MAKE) --no-print-directory _deploy
compose-check:
	@$(MAKE) --no-print-directory _config-check _compose-check
docker-build:
	@$(MAKE) --no-print-directory _docker-build
db-bootstrap:
	@$(MAKE) --no-print-directory _db-prepare
db-migrate:
	@$(MAKE) --no-print-directory _db-deploy
logging-up:
	@$(MAKE) --no-print-directory _logging-up
logging-status:
	@$(MAKE) --no-print-directory _logging-status
deployment-check:
	@$(MAKE) --no-print-directory _config-check _deployment-check
docker-network docker-ready:
	@$(MAKE) --no-print-directory _$@
# Failure diagnostics must work even when Compose/configuration setup failed.
docker-diagnostics:
	@$(MAKE) --no-print-directory _docker-diagnostics
docker-clean:
	@$(MAKE) --no-print-directory _docker-clean

.PHONY: _deploy _compose-check _docker-build _db-bootstrap _logging-up \
	_network-check _db-check _grafana-check _db-migrate _bot-replace \
	_deployment-check _bot-ownership _db-ownership _docker-network \
	_db-prepare _db-deploy _logging-status _docker-ready _docker-diagnostics _docker-clean \
	_config-check _db-config _migration-config

_config-check: _db-config _migration-config
	@test -n "$$TELEGRAM_BOT_TOKEN" || { echo 'Set TELEGRAM_BOT_TOKEN' >&2; exit 1; }

_db-config:
	@if [ -z "$$POSTGRES_URL" ]; then \
		test -n "$${DB_USER:-$$POSTGRES_USER}" && \
		test -n "$${DB_PASSWORD:-$$POSTGRES_PASSWORD}" && \
		test -n "$${DB_NAME:-$$POSTGRES_DB}" || \
		{ echo 'Set DB_USER, DB_PASSWORD, DB_NAME (or POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB)' >&2; exit 1; }; \
	fi

# JDBC is a different URL format. Never guess or discard remote URI options.
_migration-config:
	@if [ -n "$$POSTGRES_URL" ] && [ -z "$$MIGRATION_DB_URL" ]; then \
		echo 'External PostgreSQL requires MIGRATION_DB_URL (native JDBC URL); see docs/deployment.md' >&2; exit 1; \
	fi

# Explicit recursive -j invocations enforce parallel siblings AND stage barriers,
# even for plain `make deploy` (not just callers who remember `make -j`).
_deploy:
	@$(MAKE) --no-print-directory _config-check
	@$(MAKE) --no-print-directory _deployment-check _compose-check _docker-network
	@echo 'Step 1: build image / ensure local PostgreSQL / ensure local Grafana stack'
	@$(MAKE) --no-print-directory -j3 _docker-build _db-bootstrap _logging-up
	@echo 'Step 2: network / authenticated database / Grafana liveness'
	@$(MAKE) --no-print-directory -j3 _network-check _db-check _grafana-check
	@echo 'Step 3: Liquibase migration'
	@$(MAKE) --no-print-directory _db-migrate
	@echo 'Step 4: replace bot and verify startup'
	@$(MAKE) --no-print-directory _bot-replace

_compose-check:
	@$(IMAGE) config --quiet
	@$(DB) --profile tools config --quiet
	@$(BOT) config --quiet
	@if [ -z "$$GRAFANA_URL" ]; then $(GRAFANA) config --quiet; fi

_deployment-check: _bot-ownership
	@if [ -z "$$POSTGRES_URL" ]; then $(MAKE) --no-print-directory _db-ownership; fi

# Compose cannot adopt unmanaged containers. No automatic stop/remove/migration.
_bot-ownership _db-ownership:
	@set -eu; docker info >/dev/null; \
	service=bot; if [ '$@' = _db-ownership ]; then service=postgres; fi; \
	container=inanny-$$service; \
	if docker container inspect "$$container" >/dev/null 2>&1; then \
		owner=$$(docker inspect -f '{{index .Config.Labels "com.docker.compose.project"}}/{{index .Config.Labels "com.docker.compose.service"}}' "$$container"); \
		if [ "$$owner" != "inanny/$$service" ]; then \
			echo "Refusing unmanaged $$container; follow docs/deployment.md legacy migration" >&2; exit 1; \
		fi; \
	fi

_docker-network:
	@docker network inspect inanny-network >/dev/null 2>&1 || \
		docker network create inanny-network >/dev/null || \
		docker network inspect inanny-network >/dev/null

_docker-build:
	@set -eu; config=$$($(IMAGE) config --format json); \
	printf '%s\n' "$$config" | docker buildx bake -f - --load bot

_db-bootstrap:
	@if [ -z "$$POSTGRES_URL" ]; then \
		$(MAKE) --no-print-directory _db-ownership _docker-network && \
		(docker volume inspect inanny-postgres-data >/dev/null 2>&1 || docker volume create inanny-postgres-data >/dev/null) && \
		$(DB) up -d --no-deps --no-recreate postgres; \
	else echo 'Explicit POSTGRES_URL: no local PostgreSQL changes'; fi

_logging-up:
	@if [ -z "$$GRAFANA_URL" ]; then \
		$(MAKE) --no-print-directory _docker-network && \
		$(GRAFANA) up -d --no-recreate; \
	else echo 'External GRAFANA_URL: entire logging stack remains remote'; fi

_network-check:
	@docker network inspect inanny-network >/dev/null

_db-check:
	@$(DB) run --rm --no-deps db-check

# curl runs on the same network as the bot, not on the operator's laptop.
# Bearer credentials are passed via env, never expanded into host command argv.
_grafana-check:
	@docker run --rm --network inanny-network \
		--add-host host.docker.internal:host-gateway \
		-e GRAFANA_URL -e GRAFANA_HEALTH_TOKEN \
		--entrypoint sh curlimages/curl:8.12.1 -ec '\
		probe() { code=$$(curl --silent --show-error --fail --connect-timeout 5 --max-time 10 --retry 12 --retry-delay 2 --retry-all-errors --output /tmp/health --write-out "%{http_code}" "$$@"); test "$$code" = 200; }; \
		url=$${GRAFANA_URL:-http://grafana:3000}; \
		if [ -n "$${GRAFANA_HEALTH_TOKEN:-}" ]; then \
			probe -H "Authorization: Bearer $$GRAFANA_HEALTH_TOKEN" "$${url%/}/api/health"; \
		else probe "$${url%/}/api/health"; fi; \
		grep -Eq '\''"database"[[:space:]]*:[[:space:]]*"ok"'\'' /tmp/health; \
		if [ -z "$$GRAFANA_URL" ]; then probe http://loki:3100/ready; probe http://alloy:12345/-/ready; fi'

_db-migrate:
	@$(DB) run --rm --no-deps migrate

_bot-replace:
	@$(BOT) up -d --no-deps --no-build --pull never --force-recreate bot
	@$(MAKE) --no-print-directory _docker-ready

_db-prepare: _db-config
	@$(MAKE) --no-print-directory _db-bootstrap _docker-network
	@$(MAKE) --no-print-directory _db-check
_db-deploy: _migration-config
	@$(MAKE) --no-print-directory _db-prepare
	@$(MAKE) --no-print-directory _db-migrate
_logging-status:
	@if [ -z "$$GRAFANA_URL" ]; then $(GRAFANA) ps; else echo "Grafana stack is externally managed"; fi

_docker-ready:
	@set -eu; case '$(BOT_STARTUP_TIMEOUT)' in ''|*[!0-9]*) echo 'Invalid BOT_STARTUP_TIMEOUT' >&2; exit 1;; esac; \
	elapsed=0; while :; do \
		state=$$(docker inspect -f '{{.State.Status}} {{.RestartCount}}' inanny-bot); \
		if [ "$$state" != 'running 0' ]; then $(MAKE) --no-print-directory _docker-diagnostics; exit 1; fi; \
		if docker logs --tail 100 inanny-bot 2>&1 | grep -Fq 'Telegram bot Innany startup checks passed'; then \
			test "$$(docker inspect -f '{{.State.Status}} {{.RestartCount}}' inanny-bot)" = 'running 0'; \
			echo 'Bot startup confirmed'; break; \
		fi; \
		if [ "$$elapsed" -ge '$(BOT_STARTUP_TIMEOUT)' ]; then $(MAKE) --no-print-directory _docker-diagnostics; exit 1; fi; \
		sleep 1; elapsed=$$((elapsed+1)); \
	done

_docker-diagnostics:
	@for container in inanny-postgres inanny-bot; do \
		docker inspect -f 'State={{.State.Status}} ExitCode={{.State.ExitCode}} Restarts={{.RestartCount}} Mounts={{json .Mounts}} Networks={{json .NetworkSettings.Networks}}' "$$container" || true; \
		docker logs --tail 200 --since 10m "$$container" || true; \
	done

_docker-clean: _bot-ownership
	@$(BOT) rm -s -f bot
