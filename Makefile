.PHONY: build run clean test fmt lint help

BINARY_NAME=goe-mu
BINARY_UNIX=$(BINARY_NAME).exe
GO_FILES=$(shell find . -name '*.go' -type f)

help:
	@echo "Goe-Mu - NDS Emulator Build Commands"
	@echo "===================================="
	@echo "make build   - Build the emulator"
	@echo "make run     - Build and run the emulator"
	@echo "make clean   - Remove build artifacts"
	@echo "make test    - Run tests"
	@echo "make fmt     - Format all Go files"
	@echo "make lint    - Run linter"
	@echo "make deps    - Download dependencies"

build: $(GO_FILES)
	@echo "Building $(BINARY_UNIX)..."
	go build -o $(BINARY_UNIX) -v

run: build
	@echo "Running $(BINARY_UNIX)..."
	./$(BINARY_UNIX)

clean:
	@echo "Cleaning..."
	go clean
	rm -f $(BINARY_UNIX)

test:
	@echo "Running tests..."
	go test -v -cover ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	go vet ./...

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

all: clean build
	@echo "Build complete!"
