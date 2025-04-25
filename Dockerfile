# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/byc /app/byc
COPY --from=builder /app/bin/bycminer /app/bycminer

# Create data directory
RUN mkdir -p /app/data

# Expose ports
EXPOSE 8333 8334

# Set environment variables
ENV BYC_DATA_DIR=/app/data
ENV BYC_NETWORK=mainnet

# Run the application
CMD ["/app/byc"] 