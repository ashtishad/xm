.PHONY: up down down-data test lint migrate-create migrate-up migrate-down

# Helper function to extract values from app.env
define extract_env_value
$(shell grep $(1) app.env | cut -d '=' -f2-)
endef

# Variables
DB_CONN_STRING := $(if $(PGHOST),postgres://$(PGUSER):$(PGPASSWORD)@$(PGHOST):5432/$(PGDATABASE)?sslmode=disable,$(call extract_env_value,DB_CONN_STRING))

up:
	@echo "Starting Docker services..."
	@docker compose up

down:
	@echo "Stopping Docker services..."
	@docker compose down

down-data:
	@echo "Stopping Docker services and removing data..."
	@docker compose down -v --remove-orphans

test:
	@echo "Running tests with race detector..."
	@go test -race ./...

lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run ./...

migrate-create:
	@echo "Creating a new migration..."
	@migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	@echo "Running database migrations up..."
	@migrate -database "$(DB_CONN_STRING)" -path migrations up

migrate-down:
	@echo "Running database migrations down..."
	@migrate -database "$(DB_CONN_STRING)" -path migrations down
