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

build: templ-install
	@echo "Building..."
	@templ generate

	@CGO_ENABLED=1 GOOS=linux go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

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
	@go run cmd/api/main.go

db:
	@docker compose --env-file .env.local up -d db
	@sleep 1

watch/templ:
	@templ generate --watch --proxy="http://localhost:8080" --open-browser=false -v
watch/tailwind:
	@bun x tailwindcss -w -i ./input.css -o ./cmd/web/assets/css/output.css
watch/go:
	@watchexec  -r -d 200 -- APP_ENV=local go run cmd/api/main.go

# Live Reload
watch: db
	@echo "Watching..."
	make -j watch/templ watch/tailwind watch/go


.PHONY: all build run test clean watch templ-install
