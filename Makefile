# Simple Makefile for a Go project

# Build the application
all: build test
templ-install:
	@if ! command -v templ > /dev/null; then \
		read -p "Go's 'templ' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/a-h/templ/cmd/templ@latest; \
			if [ ! -x "$$(command -v templ)" ]; then \
				echo "templ installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install templ. Exiting..."; \
			exit 1; \
		fi; \
	fi

sqlc-install:
	@if ! command -v sqlc > /dev/null; then \
		read -p "Go's 'sqlc' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; \
			if [ ! -x "$$(command -v sqlc)" ]; then \
				echo "sqlc installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install sqlc. Exiting..."; \
			exit 1; \
		fi; \
	fi

goose-install:
	@if ! command -v goose > /dev/null; then \
		read -p "Go's 'goose' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/pressly/goose/v3/cmd/goose@latest; \
			if [ ! -x "$$(command -v goose)" ]; then \
				echo "goose installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install goose. Exiting..."; \
			exit 1; \
		fi; \
	fi

goose-create: goose-install
	@read -p "Enter the name of the migration: " migration; \
	goose -dir internal/database/migrations create $$migration sql

build: templ-install sqlc-install
	@echo "Building..."
	@templ generate
	@sqlc generate

	@CGO_ENABLED=1 GOOS=linux go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go
# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main


dev-run:
	@echo "Building..."
	@templ generate
	@sqlc generate
	@go run cmd/api/main.go

install-watchexec:
	@if ! command -v watchexec > /dev/null; then \
		read -p "watchexec is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			apt-get install watchexec; \
			if [ ! -x "$$(command -v watchexec)" ]; then \
				echo "watchexec installation failed"; \
			fi; \
		else \
			echo "You chose not to install watchexec. Exiting..."; \
			exit 1; \
		fi; \
	fi;


# Live Reload
watch: install-watchexec
	@watchexec -r -e go,templ,html,css,js,sql -d 1s -- make dev-run

.PHONY: all build run test clean watch templ-install sqlc-install
