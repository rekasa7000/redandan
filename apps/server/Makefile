.PHONY: dev build run tidy lint

# Hot-reload development server (requires air: go install github.com/air-verse/air@latest)
dev:
	air

# Build binary
build:
	go build -o bin/server ./cmd/main.go

# Run binary directly (no hot-reload)
run:
	go run ./cmd/main.go

# Download and tidy dependencies
tidy:
	go mod tidy

# Lint (requires golangci-lint)
lint:
	golangci-lint run ./...
