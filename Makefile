.PHONY: build test lint clean dev migrate seed

# Build variables
BINARY=vortexcms
BUILD_DIR=./bin
GO=go

# Build the application
build:
	$(GO) build -o $(BUILD_DIR)/$(BINARY) ./cmd/server

# Run all tests
test:
	$(GO) test ./... -v -count=1

# Run tests with coverage
test-cover:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

# Run in development mode
dev:
	$(GO) run ./cmd/server

# Run database migrations
migrate:
	$(GO) run ./cmd/server --migrate

# Seed the database
seed:
	$(GO) run ./cmd/server --seed

# Format code
fmt:
	$(GO) fmt ./...

# Vet code
vet:
	$(GO) vet ./...

# Generate swagger docs
swagger:
	swag init -g cmd/server/main.go -o docs/swagger

# Build Docker image
docker:
	docker build -t vortexcms:latest .

# Run with Docker Compose
docker-up:
	docker-compose up -d

# Stop Docker Compose
docker-down:
	docker-compose down
