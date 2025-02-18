FROM golang:alpine AS builder

LABEL stage=gobuilder

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .

RUN go build -ldflags "-w -s" -o /app/main cmd/inanny/main.go 


FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/main /app/main

CMD ["./main"]