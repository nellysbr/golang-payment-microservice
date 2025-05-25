.PHONY: help build run test clean docker-build docker-run docker-compose-up docker-compose-down deps lint fmt vet

# Variables
APP_NAME=payment-microservice
DOCKER_IMAGE=payment-microservice:latest
DOCKER_COMPOSE_FILE=docker-compose.yml

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
deps: ## Download dependencies
	go mod download
	go mod tidy

build: ## Build the application
	go build -o bin/$(APP_NAME) cmd/main.go

run: ## Run the application locally
	go run cmd/main.go

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -cover ./...

test-coverage-html: ## Generate HTML coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality
lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

# Docker
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	docker run -p 8080:8080 -p 2112:2112 $(DOCKER_IMAGE)

# Docker Compose
docker-compose-up: ## Start all services with Docker Compose
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

docker-compose-down: ## Stop all services
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

docker-compose-logs: ## Show logs from all services
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

docker-compose-restart: ## Restart all services
	docker-compose -f $(DOCKER_COMPOSE_FILE) restart

# Database
db-migrate: ## Run database migrations
	docker-compose exec postgres psql -U postgres -d payment_db -f /docker-entrypoint-initdb.d/001_create_tables.sql

db-reset: ## Reset database (WARNING: This will delete all data)
	docker-compose down -v
	docker-compose up -d postgres
	sleep 5
	make db-migrate

# Monitoring
prometheus: ## Open Prometheus in browser
	@echo "Opening Prometheus at http://localhost:9090"
	@which xdg-open > /dev/null && xdg-open http://localhost:9090 || echo "Please open http://localhost:9090 in your browser"

grafana: ## Open Grafana in browser
	@echo "Opening Grafana at http://localhost:3000 (admin/admin)"
	@which xdg-open > /dev/null && xdg-open http://localhost:3000 || echo "Please open http://localhost:3000 in your browser"

# API Testing
test-api: ## Test API endpoints
	@echo "Testing health endpoint..."
	curl -s http://localhost:8080/health | jq .
	@echo "\nTesting metrics endpoint..."
	curl -s http://localhost:2112/metrics | head -10

test-payment: ## Create a test payment
	curl -X POST http://localhost:8080/api/v1/payments \
		-H "Content-Type: application/json" \
		-d '{"card_number":"1234567890123456","card_holder":"Test User","expiry_month":12,"expiry_year":2025,"cvv":"123","amount":100.50,"currency":"BRL","merchant_id":"test-merchant"}' | jq .

# Cleanup
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker system prune -f

clean-all: ## Clean everything including Docker volumes
	make clean
	docker-compose down -v
	docker system prune -a -f

# Development workflow
dev-setup: ## Setup development environment
	make deps
	make docker-compose-up
	@echo "Waiting for services to start..."
	sleep 10
	make test-api

dev-restart: ## Restart development environment
	make docker-compose-down
	make docker-compose-up
	@echo "Waiting for services to start..."
	sleep 10

# Production
prod-build: ## Build for production
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(APP_NAME) cmd/main.go

# Benchmarks
benchmark: ## Run benchmarks
	go test -bench=. -benchmem ./...

# Security
security-scan: ## Run security scan with gosec
	gosec ./...

# Generate
generate: ## Run go generate
	go generate ./...

# Install tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest 