FROM golang:1.21-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o tormenta-bot ./cmd/bot/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/tormenta-bot .

# Create assets directory (images are auto-generated on first run)
RUN mkdir -p /app/assets/images

ENV ASSETS_DIR=/app/assets/images

# Persist generated images between container restarts
VOLUME ["/app/assets/images"]

CMD ["./tormenta-bot"]
