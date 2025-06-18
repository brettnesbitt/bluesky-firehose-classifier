.PHONY: lint test benchmark build docker-build docker-up docker-down clean

# Set def linttarget
default: build

# Define Go version
GO_VERSION := 1.21

# Define build flags
BUILD_FLAGS := -ldflags "-X 'main.Version=$(VERSION)'"

# Set the version
#VERSION := $(shell git describe --tags --dirty)

# Set our golangci-lint path
GOLANGCI_LINT_PATH := ./$(go env GOPATH)/bin/golangci-lint

# Set our docker image name
IMAGE_NAME := bluesky-firehose-classifier:latest

# docker compose profile
COMPOSE_PROFILES ?= $(shell ./scripts/get_docker_profile.sh)

lint: 
	@echo "Linting Code..."
	@golangci-lint run ./...

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run benchmarks
benchmark:
	@echo "Running tests..."
	@go test -v ./... -bench=. -benchmem

# Build the application
build: lint test benchmark
	@echo "Building all binaries..."
	@go mod tidy
	@echo "Building server..."
	@CGO_ENABLED=0 go build -o bin/server -v $(BUILD_FLAGS) ./cmd/server
	@echo "Building wss..."
	@CGO_ENABLED=0 go build -o bin/wss -v $(BUILD_FLAGS) ./cmd/wss

build-app-only:
	@echo "Building all binaries..."
	@go mod tidy
	@echo "Building server..."
	@CGO_ENABLED=0 go build -o bin/server -v $(BUILD_FLAGS) ./cmd/server
	@echo "Building wss..."
	@CGO_ENABLED=0 go build -o bin/wss -v $(BUILD_FLAGS) ./cmd/wss

# Build the docker image
docker-build: build-app-only
	@COMPOSE_PROFILES=$(COMPOSE_PROFILES) docker compose build

docker-up: docker-build
	@COMPOSE_PROFILES=$(COMPOSE_PROFILES) docker compose up -d

docker-down:
	@COMPOSE_PROFILES=$(COMPOSE_PROFILES) docker compose down

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf bin/*
	@docker rmi $(IMAGE_NAME) || true 
