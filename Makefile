.PHONY: run watch test lint up down down-data migrate-create migrate-up migrate-down \
         docker-build docker-run docker-stop docker-rebuild-run


# Helper function to extract values from app.env
define extract_env_value
$(shell grep $(1) app.env | cut -d '=' -f2-)
endef

# Variables
DB_CONN_STRING := $(if $(PGHOST),postgres://$(PGUSER):$(PGPASSWORD)@$(PGHOST):5432/$(PGDATABASE)?sslmode=disable,$(call extract_env_value,DB_CONN_STRING))

# Application commands (run, test, lint or live-reload the app)
run:
	@echo "Running the application..."
	@go run main.go

watch:
	@if command -v air > /dev/null; then \
		echo "Watching with air..."; \
		air; \
	else \
		read -p "Go's 'air' is not installed. Install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			echo "Installing air..."; \
			go install github.com/cosmtrek/air@latest; \
			echo "Watching with air..."; \
			air; \
		else \
			echo "Air installation declined. Exiting..."; \
			exit 1; \
		fi; \
	fi

test:
	@echo "Running tests with race detector..."
	@go test -race ./...

lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run ./...

# Docker compose commands
up:
	@echo "Starting Docker services..."
	@docker compose up -d

down:
	@echo "Stopping Docker services..."
	@docker compose down

down-data:
	@echo "Stopping Docker services and removing data..."
	@docker compose down -v --remove-orphans

# Database commands
migrate-create:
	@echo "Creating a new migration..."
	@migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	@echo "Running database migrations up..."
	@migrate -database "$(DB_CONN_STRING)" -path migrations up

migrate-down:
	@echo "Running database migrations down..."
	@migrate -database "$(DB_CONN_STRING)" -path migrations down

# Docker application commands
docker-build:
	@echo "Building Docker image..."
	@docker build -t xm:latest .

docker-run:
	@echo "Running application in Docker container..."
	@docker run --name app --network xm_network \
		-e DB_CONN_STRING="$(call extract_env_value,DB_CONN_STRING)" \
		-e SERVER_ADDRESS="$(call extract_env_value,SERVER_ADDRESS)" \
		-p 8080:8080 xm:latest

docker-stop:
	@echo "Stopping and removing Docker container..."
	@docker stop app || true
	@docker rm app || true

docker-rebuild-run: docker-stop docker-build docker-run
