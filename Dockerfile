#######################
#   Bun dependencies   #
#######################
FROM oven/bun:1 AS bun-deps

WORKDIR /usr/src/app

COPY package.json bun.lock ./

RUN bun install

################
#   Tailwind   #
################
FROM oven/bun:1 AS tailwind

WORKDIR /usr/src/app

# Copy dependencies
COPY --from=bun-deps /usr/src/app/node_modules ./node_modules

# Copy all templ files
COPY . .

# Build Tailwind CSS
RUN bun build:tailwind

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

# Copy static assets and CSS
COPY --from=tailwind /usr/src/app/output.css /app/cmd/web/assets/css/output.css

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

# Expose application port
EXPOSE 8080

# Run the application
CMD ["./main"]
