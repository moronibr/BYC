.PHONY: all build test clean lint docker-build docker-run

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=byc
BINARY_UNIX=$(BINARY_NAME)_unix

# Docker parameters
DOCKER_IMAGE=byc
DOCKER_TAG=latest

all: test build

build:
	$(GOBUILD) -o bin/byc cmd/youngchain/main.go
	$(GOBUILD) -o bin/bycminer cmd/bycminer/main.go

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f bin/bycminer

lint:
	golangci-lint run

# Docker targets
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run:
	docker run -p 8333:8333 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Development tools
deps:
	$(GOGET) -v -t -d ./...

# Generate mocks for testing
generate:
	mockgen -source=internal/network/server.go -destination=internal/network/mocks/server_mock.go
	mockgen -source=internal/core/mining/mining.go -destination=internal/core/mining/mocks/mining_mock.go

# Run all checks
check: lint test 