.PHONY: build test run clean docker-build docker-up docker-down

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
