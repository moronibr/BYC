# Build stage
FROM golang:1.16 AS builder

# Install build dependencies
RUN apt-get update && \
    apt-get install -y libleveldb-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Final stage
FROM ubuntu:20.04

# Install runtime dependencies
RUN apt-get update && \
    apt-get install -y libleveldb-dev ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -u 1000 byc

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/byc /app/byc

# Copy configuration files
COPY --from=builder /app/config /app/config

# Set ownership
RUN chown -R byc:byc /app

# Switch to non-root user
USER byc

# Expose ports
EXPOSE 8000

# Set entrypoint
ENTRYPOINT ["/app/byc"]

# Default command
CMD ["node", "start"] 