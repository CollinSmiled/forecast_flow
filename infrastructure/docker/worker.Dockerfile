FROM golang:1.27.1-alpine AS build

WORKDIR /src

ARG APP_NAME

COPY go.mod go.sum ./
RUN go mod download

COPY apps ./apps
COPY internal ./internal

RUN test -n "${APP_NAME}" \
    && CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/app \
        ./apps/${APP_NAME}

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app

COPY --from=build /out/app /usr/local/bin/app

USER app

ENTRYPOINT ["/usr/local/bin/app"]
