#!/bin/sh

if [ -z "$1" ]; then
    echo "Error: No argument provided."
    echo "Usage: $0 {run|build|rab}"
    exit 1
fi

check_env() {
    if [ -z "${TELEGRAM_BOT_TOKEN}" ]; then
        echo "TELEGRAM_BOT_TOKEN is not set."
        exit 1
    fi
}

run() {
    docker run --rm -d \
        -e TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN \
        --name inanny-bot \
        inanny-bot
}

build() {
    docker build -t inanny-bot .
}

case "$1" in
    run)
        echo "Running Docker container..."
        check_env
        run
        ;;
    build)
        echo "Building Docker image..."
        build
        ;;
    bar)
        echo "Rebuilding and running Docker container..."
        check_env
        build
        run
        ;;
    *)
        echo "Error: Invalid argument."
        echo "Usage: $0 {run|build|rab}"
        exit 1
        ;;
esac
