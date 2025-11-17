.PHONY: build run test clean docker docker-up docker-down help

# Variables
BINARY_NAME=gateway
DOCKER_IMAGE=ai-gateway
DOCKER_TAG=latest

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building..."
	go build -o $(BINARY_NAME) ./cmd/gateway
	@echo "Build complete: $(BINARY_NAME)"

run: ## Run the application
	@echo "Running..."
	go run ./cmd/gateway/main.go

test: ## Run tests
	@echo "Running tests..."
	go test ./... -v -cover

test-short: ## Run tests without verbose output
	@echo "Running tests..."
	go test ./... -cover

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean
	@echo "Clean complete"

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

lint: fmt vet ## Run formatters and linters

docker: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-up: ## Start services with docker-compose
	@echo "Starting services..."
	docker-compose up -d
	@echo "Services started. Use 'make docker-logs' to view logs"

docker-down: ## Stop services with docker-compose
	@echo "Stopping services..."
	docker-compose down
	@echo "Services stopped"

docker-logs: ## View docker-compose logs
	docker-compose logs -f

docker-clean: docker-down ## Clean docker containers and images
	@echo "Cleaning Docker resources..."
	docker-compose down -v
	@echo "Docker resources cleaned"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	@echo "Dependencies downloaded"

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	go mod tidy
	@echo "go.mod tidied"

api-test: ## Run API test script (requires service to be running)
	@echo "Running API tests..."
	@bash scripts/test_api.sh

install: ## Install the binary
	@echo "Installing..."
	go install ./cmd/gateway
	@echo "Installed"

all: clean lint test build ## Clean, lint, test and build
