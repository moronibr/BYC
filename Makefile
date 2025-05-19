.PHONY: all build test clean lint run

# Variables
BINARY_NAME=byc
BUILD_DIR=bin
VERSION=1.0.0
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

all: clean build

build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	@go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} cmd/byc/main.go

test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

lint:
	@echo "Running linter..."
	@golangci-lint run

clean:
	@echo "Cleaning..."
	@rm -rf ${BUILD_DIR}
	@go clean

run: build
	@echo "Running ${BINARY_NAME}..."
	@./${BUILD_DIR}/${BINARY_NAME}

deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Release
release: clean build
	@echo "Creating release..."
	@mkdir -p release
	@tar -czf release/${BINARY_NAME}-${VERSION}-linux-amd64.tar.gz -C ${BUILD_DIR} ${BINARY_NAME}
	@tar -czf release/${BINARY_NAME}-${VERSION}-darwin-amd64.tar.gz -C ${BUILD_DIR} ${BINARY_NAME}
	@tar -czf release/${BINARY_NAME}-${VERSION}-windows-amd64.tar.gz -C ${BUILD_DIR} ${BINARY_NAME}.exe

help:
	@echo "Available targets:"
	@echo "  all            - Clean and build the project"
	@echo "  build          - Build the project"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  run            - Build and run the project"
	@echo "  deps           - Install dependencies"
	@echo "  install-tools  - Install development tools"
	@echo "  release        - Create release packages"
	@echo "  help           - Show this help message" 