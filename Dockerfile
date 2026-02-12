FROM golang:1.24-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG COMMIT
ARG BRANCH

RUN COMMIT=${COMMIT:-$(git describe --always --dirty --abbrev=7)} && \
    BRANCH=${BRANCH:-$(git rev-parse --abbrev-ref HEAD)} && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
    -X github.com/jus1d/kypidbot/internal/version.Commit=$COMMIT \
    -X github.com/jus1d/kypidbot/internal/version.Branch=$BRANCH" \
    -o /out/kypidbot ./cmd/bot && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /out/migrate ./cmd/migrate

# lightweight docker container with binaries only
FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache tzdata

COPY --from=builder /out/kypidbot .
COPY --from=builder /out/migrate .

COPY --from=builder /build/messages ./messages
COPY --from=builder /build/migrations ./migrations

COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

RUN addgroup -S app && adduser -S app -G app
USER app

ENTRYPOINT ["/app/docker-entrypoint.sh"]
