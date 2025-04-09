.PHONY: help dev up down build test clean proto logs migrate cockroach-shell

# Default target
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Start services and run the Go server locally (depends on `up` command)
serve: ## Run the Go server locally
up: ## Start all Docker services
down: ## Stop all Docker services
build: ## Build the Go binary
test: ## Run Go tests
clean: ## Remove build artifacts and stop services
proto: ## Generate protobuf and gRPC code
logs: ## Tail logs from all services
migrate: ## Apply latest SpiceDB migrations
migrate-docker: ## Apply latest SpiceDB migrations using Docker
cockroach-shell: ## Open a CockroachDB SQL shell

# Development commands
dev: up serve

serve:
	@echo "Starting server..."
	@go run cmd/server/main.go --config="default.yaml" --env=development 

up:
	@echo "Starting services..."
	@docker-compose up -d

down:
	@echo "Stopping services..."
	@docker-compose down

build:
	@echo "Building service..."
	@go build -o main cmd/server/main.go

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning up..."
	@rm -f main
	@docker-compose down -v

proto:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/authorization.proto

logs:
	@echo "Showing logs..."
	@docker-compose logs -f

migrate: ## Apply latest SpiceDB migrations using Go migration tool
	@echo "Applying SpiceDB migrations..."
	@go run cmd/spicedb-migrate/main.go --config="default.yaml" --env=development

migrate-docker:
	@echo "Applying SpiceDB schema..."
	@docker-compose exec spicedb spicedb migrate head

cockroach-shell:
	@echo "Opening CockroachDB shell..."
	@docker-compose exec cockroachdb cockroach sql --insecure 