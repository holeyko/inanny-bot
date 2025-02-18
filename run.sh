#!/bin/sh

if [ -z "$1" ]; then
    echo "Error: No argument provided."
    echo "Usage: $0 {run|build|rab}"
    exit 1
fi

check_env() {
    echo "Check env variables..."

    if [ -z "${TELEGRAM_BOT_TOKEN}" ]; then
        echo "TELEGRAM_BOT_TOKEN is not set."
        exit 1
    fi
}

run() {
    echo "Run docker container..."

    docker run -d \
        -e TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN \
        --name inanny-bot \
        inanny-bot
}

build() {
    echo "Build docker image..."

    docker build -t inanny-bot .
}

clean() {
    echo "Clean previous docker container..."

    docker stop inanny-bot >> /dev/null
    docker rm inanny-bot >> /dev/null
}

case "$1" in
    run)
        check_env
        clean
        run
        ;;
    build)
        build
        ;;
    bar)
        check_env
        build
        clean
        run
        ;;
    *)
        echo "Error: Invalid argument."
        echo "Usage: $0 {run|build|rab}"
        exit 1
        ;;
esac
