# Stage 1: build both server and seed binaries
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o seed   ./cmd/seed/main.go

# ── Server runtime ────────────────────────────────────────────
FROM alpine:3.20 AS server

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]

# ── Seed runtime (one-shot, exits after seeding) ──────────────
FROM alpine:3.20 AS seed

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/seed .

CMD ["./seed"]
