FROM golang:1.24-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git # purely for baking commit and branch into executable

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o migrate ./cmd/migrate
RUN CGO_ENABLED=0 go build -a -ldflags "-w -s \
    -X github.com/jus1d/kypidbot/internal/version.Commit=$(git describe --always --dirty --abbrev=7) \
    -X github.com/jus1d/kypidbot/internal/version.Branch=$(git rev-parse --abbrev-ref HEAD)" \
    -o kypidbot ./cmd/bot

# lightweight docker container with binaries only
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /build/kypidbot .
COPY --from=builder /build/migrate .
COPY --from=builder /build/messages ./messages
COPY --from=builder /build/migrations ./migrations

RUN apk add --no-cache tzdata

CMD ["sh", "-c", "./migrate up && exec ./kypidbot"]
