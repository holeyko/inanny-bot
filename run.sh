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

docbuild() {
    echo "Build docker image..."

    docker build -t inanny-bot .
}

run() {
    go run cmd/inanny/main.go
}

docclean() {
    echo "Clean previous docker container..."

    docker stop inanny-bot >> /dev/null
    docker rm inanny-bot >> /dev/null
}

docrun() {
    docclean
    echo "Run docker container..."

    docker run -d \
        -e TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN \
        -e DB_HOST=$DB_HOST \
        -e DB_USER=$DB_USER \
        -e DB_PASSWORD=$DB_PASSWORD \
        -e DB_NAME=$DB_NAME \
        --name inanny-bot \
        inanny-bot
}

case "$1" in
    run)
        check_env
        run
        ;;
    docrun)
        docbuild
        check_env
        docrun
        ;;
    docbuild)
        docbuild
        ;;
    *)
        echo "Error: Invalid argument."
        echo "Usage: $0 {run|build|rab}"
        exit 1
        ;;
esac
