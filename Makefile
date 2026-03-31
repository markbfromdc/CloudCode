.PHONY: build test run clean docker-build docker-up docker-down fmt vet coverage test-frontend test-all evaluate

# Build the API server binary.
build:
	go build -ldflags="-w -s" -o cloudide-api ./cmd/server

# Run all unit tests.
test:
	go test -v -race ./...

# Run the API server locally.
run: build
	JWT_SECRET=dev-secret ./cloudide-api

# Remove build artifacts.
clean:
	rm -f cloudide-api

# Build all Docker images.
docker-build:
	docker build -t cloudide-api:latest -f Dockerfile.api .
	docker build -t cloudide-workspace:latest ./workspace

# Start all services with Docker Compose.
docker-up: docker-build
	docker compose up -d

# Stop all services.
docker-down:
	docker compose down -v

# Run linting.
lint:
	golangci-lint run ./...

# Format all Go source files.
fmt:
	go fmt ./...

# Run Go vet on all packages.
vet:
	go vet ./...

# Generate test coverage report.
coverage:
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out

# Run frontend tests.
test-frontend:
	cd frontend && npm test

# Run all backend and frontend tests.
test-all: test test-frontend

# Run project completion evaluation.
evaluate:
	go run ./cmd/evaluate -dir . -format both -output evaluation-report.json
