DOCKER_IMAGE = inanny-bot
DOCKER_CONTAINER = inanny-bot
MAIN_PATH = ./cmd/inanny/main.go
BUILD_DIR = ./build
APP = app

DOCKER_VERSION ?= latest


.PHONY: clean build run docker-clean docker-build docker-run bar


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
	@docker stop $(DOCKER_CONTAINER) > /dev/null || true
	@docker rm $(DOCKER_CONTAINER) > /dev/null || true

docker-build:
	@echo "Build docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_VERSION) .

docker-run: docker-clean
	@echo "Run docker container..."
	@docker run -d \
		-e TELEGRAM_BOT_TOKEN=$(TELEGRAM_BOT_TOKEN) \
		-e DB_HOST=$(DB_HOST) \
		-e DB_USER=$(DB_USER) \
		-e DB_PASSWORD=$(DB_PASSWORD) \
		-e DB_NAME=$(DB_NAME) \
		--name $(DOCKER_CONTAINER) \
		$(DOCKER_IMAGE):$(DOCKER_VERSION)

docker-bar: docker-build docker-run