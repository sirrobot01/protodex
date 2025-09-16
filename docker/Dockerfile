# Build stage
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-s -w -X github.com/sirrobot01/protodex/internal/cli.version=docker" \
    -o protodex ./cmd/protodex

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 protodex && \
    adduser -D -s /bin/sh -u 1000 -G protodex protodex

# Set working directory
WORKDIR /home/protodex

# Copy the binary from builder stage
COPY --from=builder /app/protodex .

# Change ownership
RUN chown protodex:protodex protodex

# Switch to non-root user
USER protodex

# Expose port
EXPOSE 8080

# Default command
CMD ["./protodex", "serve", "--port", "8080"]