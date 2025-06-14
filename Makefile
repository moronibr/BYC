# BYC Makefile

# Variables
BINARY_NAME=byc
NODE_NAME=byc-node
BUILD_DIR=bin
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

# Default target
all: clean build

# Build the CLI application
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	@go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ./cmd/byc/...

# Build the node
build-node:
	@echo "Building ${NODE_NAME}..."
	@mkdir -p ${BUILD_DIR}
	@go build ${LDFLAGS} -o ${BUILD_DIR}/${NODE_NAME} ./cmd/byc-node/...

# Build both CLI and node
build-all: clean build build-node

# Build for multiple platforms
release: clean
	@echo "Building for multiple platforms..."
	@mkdir -p ${BUILD_DIR}
	# Build CLI
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 ./cmd/byc/...
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-windows-amd64.exe ./cmd/byc/...
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64 ./cmd/byc/...
	# Build Node
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${NODE_NAME}-linux-amd64 ./cmd/byc-node/...
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${NODE_NAME}-windows-amd64.exe ./cmd/byc-node/...
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${NODE_NAME}-darwin-amd64 ./cmd/byc-node/...

# Clean build directory
clean:
	@echo "Cleaning..."
	@rm -rf ${BUILD_DIR}

# Run the CLI application
run: build
	@echo "Running ${BINARY_NAME}..."
	@./${BUILD_DIR}/${BINARY_NAME}

# Run the node
run-node: build-node
	@echo "Running ${NODE_NAME}..."
	@./${BUILD_DIR}/${NODE_NAME}

# Install both CLI and node
install: build build-node
	@echo "Installing ${BINARY_NAME} and ${NODE_NAME}..."
	@cp ${BUILD_DIR}/${BINARY_NAME} /usr/local/bin/
	@cp ${BUILD_DIR}/${NODE_NAME} /usr/local/bin/

# Uninstall both CLI and node
uninstall:
	@echo "Uninstalling ${BINARY_NAME} and ${NODE_NAME}..."
	@rm -f /usr/local/bin/${BINARY_NAME}
	@rm -f /usr/local/bin/${NODE_NAME}

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Help
help:
	@echo "Available targets:"
	@echo "  all        - Clean and build both applications"
	@echo "  build      - Build the CLI application"
	@echo "  build-node - Build the node application"
	@echo "  build-all  - Build both CLI and node"
	@echo "  release    - Build for multiple platforms"
	@echo "  clean      - Clean build directory"
	@echo "  run        - Run the CLI application"
	@echo "  run-node   - Run the node application"
	@echo "  install    - Install both applications"
	@echo "  uninstall  - Uninstall both applications"
	@echo "  test       - Run tests"
	@echo "  help       - Show this help message"

.PHONY: all build build-node build-all release clean run run-node install uninstall test help 