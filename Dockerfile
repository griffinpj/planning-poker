# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies (GCC for CGO/SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

# Install sqlite3 and ca-certificates
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy templates and static files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/migrations ./migrations

# Create directory for database
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV DB_PATH=/app/data/poker.db

# Run the application
CMD ["./main"]