.PHONY: help run build test lint docker clean fetch-data

# Default target
help:
	@echo "Available targets:"
	@echo "  make run        - Run the API locally"
	@echo "  make build      - Build the binary"
	@echo "  make test       - Run tests"
	@echo "  make lint       - Run linters"
	@echo "  make docker     - Build Docker image"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make fetch-data - Fetch sample PubMed data"

# Run locally
run:
	@echo "Starting API server..."
	@export PORT=8080 && \
	export DATA_PATH=./data/sample_100_pubmed.jsonl && \
	export LOG_LEVEL=info && \
	go run ./cmd/api

# Build binary
build:
	@echo "Building binary..."
	@CGO_ENABLED=1 go build -o bin/pubmed-api ./cmd/api

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run linters
lint:
	@echo "Running linters..."
	@go vet ./...
	@if command -v staticcheck > /dev/null; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed, skipping..."; \
	fi

# Build Docker image
docker:
	@echo "Building Docker image..."
	@docker build -t pubmed-api:local .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --rm \
		-e DATA_PATH=/app/data/sample_100_pubmed.jsonl \
		-e LOG_LEVEL=info \
		pubmed-api:local

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f *.db *.sqlite
	@go clean

# Fetch sample data (requires internet)
fetch-data:
	@echo "Fetching sample PubMed data..."
	@go run ./scripts/fetch_pubmed_data.go

