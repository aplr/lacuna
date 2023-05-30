FROM golang:1.20-alpine as builder

RUN apk add --no-cache --update build-base

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download

COPY . /app/

ARG VERSION
ENV VERSION=${VERSION}

RUN make build

FROM alpine:3.18

ENV PUBSUB_EMULATOR_HOST=127.0.0.1:8085

COPY --from=builder /app/bin/lacuna /usr/local/bin/lacuna

CMD ["lacuna", "daemon", "-vvv"]
