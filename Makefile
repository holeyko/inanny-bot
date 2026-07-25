DOCKER_IMAGE = inanny-bot
DOCKER_CONTAINER = inanny-bot
DOCKER_NETWORK = inanny-network
DB_CONTAINER = inanny-postgres
DB_VOLUME = inanny-postgres-data
MAIN_PATH = ./cmd/inanny/main.go
BUILD_DIR = ./build
APP = app

IMAGE_VERSION ?= latest
DB_PORT ?= 5432
DB_HOST ?= $(DB_CONTAINER)
POSTGRES_IMAGE ?= postgres:16
LIQUIBASE_IMAGE ?= liquibase/liquibase:4.31

export TELEGRAM_BOT_TOKEN
export DEBUG
export DB_USER
export DB_PASSWORD
export DB_NAME
export POSTGRES_USER = $(DB_USER)
export POSTGRES_PASSWORD = $(DB_PASSWORD)
export POSTGRES_DB = $(DB_NAME)


.PHONY: clean build run docker-network db-bootstrap docker-clean docker-build docker-run docker-bar deploy db-migrate bar


clean:
	go clean

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

bar: build run

docker-clean:
	@echo "Clean previous docker container..."

	@-docker stop $(DOCKER_CONTAINER) > /dev/null || true
	@-docker rm $(DOCKER_CONTAINER) > /dev/null || true

docker-network:
	@echo "Ensure docker network exists..."

	@docker network inspect $(DOCKER_NETWORK) > /dev/null 2>&1 || \
		docker network create $(DOCKER_NETWORK) > /dev/null

db-bootstrap: docker-network
	@echo "Ensure PostgreSQL container is running..."

	@docker volume inspect $(DB_VOLUME) > /dev/null 2>&1 || \
		docker volume create $(DB_VOLUME) > /dev/null
	@status=$$(docker inspect -f '{{.State.Status}}' $(DB_CONTAINER) 2> /dev/null || true); \
	if [ "$$status" = "running" ]; then \
		echo "PostgreSQL container is already running; leaving it untouched."; \
	else \
		if [ -n "$$status" ]; then \
			echo "PostgreSQL container status is '$$status'; recreating container and preserving volume."; \
			docker rm -f $(DB_CONTAINER) > /dev/null; \
		fi; \
		echo "Creating PostgreSQL container with persistent volume."; \
		docker pull $(POSTGRES_IMAGE) > /dev/null; \
		docker run -d \
			--name $(DB_CONTAINER) \
			--restart unless-stopped \
			--network $(DOCKER_NETWORK) \
			-v $(DB_VOLUME):/var/lib/postgresql/data \
			-p $(DB_PORT):5432 \
			-e POSTGRES_USER \
			-e POSTGRES_PASSWORD \
			-e POSTGRES_DB \
			$(POSTGRES_IMAGE) > /dev/null; \
	fi
	@echo "Wait for PostgreSQL to accept connections..."
	@timeout=60; elapsed=0; \
	until docker exec $(DB_CONTAINER) pg_isready -U $(DB_USER) -d $(DB_NAME) > /dev/null 2>&1; do \
		if [ $$elapsed -ge $$timeout ]; then \
			echo "PostgreSQL did not become ready after $${timeout}s."; \
			echo "Docker containers:"; \
			docker ps -a; \
			echo "PostgreSQL container details:"; \
			docker inspect $(DB_CONTAINER) || true; \
			echo "PostgreSQL logs:"; \
			docker logs --tail 200 $(DB_CONTAINER) || true; \
			exit 1; \
		fi; \
		sleep 2; \
		elapsed=$$((elapsed + 2)); \
	done

docker-build:
	@echo "Build docker image..."

	docker build -t $(DOCKER_IMAGE):$(IMAGE_VERSION) .

docker-run: docker-network
	@echo "Run docker container..."

	@docker run -d \
		--restart unless-stopped \
		--network $(DOCKER_NETWORK) \
		-e TELEGRAM_BOT_TOKEN \
		-e DEBUG \
		-e DB_HOST=$(DB_HOST) \
		-e DB_USER \
		-e DB_PASSWORD \
		-e DB_NAME \
		--name $(DOCKER_CONTAINER) \
		$(DOCKER_IMAGE):$(IMAGE_VERSION)
	@sleep 5
	@echo "Bot container status after startup:"
	@docker ps -a --filter name=$(DOCKER_CONTAINER)
	@echo "Bot container logs after startup:"
	@docker logs --tail 100 $(DOCKER_CONTAINER) || true

docker-bar: docker-build docker-clean docker-run
	
db-migrate: docker-network
	@LIQUIBASE_COMMAND_URL="jdbc:postgresql://$(DB_CONTAINER):5432/$(DB_NAME)" \
	LIQUIBASE_COMMAND_USERNAME="$(DB_USER)" \
	LIQUIBASE_COMMAND_PASSWORD="$(DB_PASSWORD)" \
	docker run --rm \
		--network "$(DOCKER_NETWORK)" \
		--mount type=bind,src="$(CURDIR)/migrate",dst=/liquibase/changelog,readonly \
		-e LIQUIBASE_COMMAND_URL \
		-e LIQUIBASE_COMMAND_USERNAME \
		-e LIQUIBASE_COMMAND_PASSWORD \
		$(LIQUIBASE_IMAGE) \
		--changelog-file=changelog.xml update

deploy:
	$(MAKE) docker-build
	$(MAKE) db-bootstrap
	$(MAKE) db-migrate
	$(MAKE) docker-clean
	$(MAKE) docker-run
