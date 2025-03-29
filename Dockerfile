################
#   Tailwind   #
################
FROM oven/bun:1 AS tailwind

WORKDIR /usr/src/app

# Copy all templ files
COPY ./**/*.templ ./
COPY ./cmd/web/assets/css/input.css ./input.css

# Install Tailwind CSS
RUN bun init -y
RUN bun add tailwindcss @tailwindcss/cli@latest

# Run Tailwind CSS

RUN bun x @tailwindcss/cli@latest -i ./input.css -o ./output.css

################
# Dependencies #
################
FROM golang:1.24-alpine AS builder

# Install required system dependencies
RUN apk add --no-cache gcc musl-dev make git curl

# Install Go tools
RUN go install github.com/a-h/templ/cmd/templ@latest

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Run templ generation
RUN templ generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

#############
# Main part #
#############
FROM alpine:latest

# Install CA certificates for HTTPS connections
RUN apk add --no-cache ca-certificates

# Set working directory
WORKDIR /app

# Copy compiled binary
COPY --from=builder /app/main .

# Copy static assets and CSS
COPY --from=builder /app/cmd/web/assets /app/cmd/web/assets
COPY --from=tailwind /usr/src/app/output.css /app/cmd/web/assets/css/output.css

# Expose application port
EXPOSE 8080

# Run the application
CMD ["./main"]
