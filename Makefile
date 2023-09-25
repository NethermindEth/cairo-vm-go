.PHONY: build clean test help format staticcheck pre-commit

BINARY_DIR := bin
BINARY_NAME := cairo-vm

default: help

help:
	@echo "This makefile allos the following commands"
	@echo "  make build       - compile the source code"
	@echo "  make clean       - remove binary files"
	@echo "  make test        - run tests"
	@echo "  make help        - show this help message"

build:
	@echo "Building..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(BINARY_DIR)/$(BINARY_NAME) cmd/cli/main.go
	@if [ $$? -eq 0 ]; then \
		echo "Build completed succesfully!"; \
	else \
		echo "Build failed."; \
		exit 1; \
	fi

clean:
	@echo "Cleaning up..."
	@rm -rf $(BINARY_DIR)

unit:
	@echo "Running unit tests..."
	@go test ./pkg/...

integration:
	@echo "Running integration tests..."
	@go test ./integration_tests/...

testall:
	@echo "Running all tests..."
	@go test ./...

format:
	@gofumpt -l -w .

staticcheck:
	@staticcheck ./...

pre-commit: format staticcheck build test clean
