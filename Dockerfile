################
# Dependencies #
################

FROM golang:1.24-alpine AS builder

# Install required system dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev make git curl

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Run templ and sqlc generation
RUN templ generate
RUN sqlc generate

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o main cmd/api/main.go

#############
# Main part #
#############

FROM alpine:latest

# Install runtime dependencies for SQLite
RUN apk add --no-cache sqlite-libs ca-certificates

# Set working directory
WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

# Create directory for SQLite database
RUN mkdir -p /app/data

# Run the application
CMD ["./main"]
