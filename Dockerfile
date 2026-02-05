FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o kypidbot ./cmd/bot
RUN CGO_ENABLED=0 go build -o migrate ./cmd/migrate

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /build/kypidbot .
COPY --from=builder /build/migrate .
COPY --from=builder /build/messages.yaml .
COPY --from=builder /build/migrations ./migrations
CMD ["sh", "-c", "./migrate up && ./kypidbot"]
