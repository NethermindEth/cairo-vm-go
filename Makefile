.PHONY: lint build clean test help format staticcheck pre-commit

GOPATH_DIR :=`go env GOPATH`
BINARY_DIR := bin
BINARY_NAME := cairo-vm

TEST := "."

default: help

help:
	@echo "This makefile allows the following commands"
	@echo "  make build           - compile the source code"
	@echo "  make clean           - remove binary files"
	@echo "  make unit            - run unit tests"
	@echo "  make integration     - run integration tests"
	@echo "  make testall         - run all tests"
	@echo "  make bench           - benchmark all tests"
	@echo "  make help            - show this help message"

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
	@$(MAKE) build
	@if [ $$? -eq 0 ]; then \
		go test ./integration_tests/... -v; \
	else \
		echo "Integration tests were not run"; \
		exit 1; \
	fi

testall:
	@echo "Running all tests..."
	@$(MAKE) build
	@go test ./...

bench:
	@echo "Running benchmarks..."
	@go run scripts/benchmark.go --pkg=${PKG_NAME} --test=${TEST}

# Use the same version of the golangci-lint as in our CI linting config.
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3
	golangci-lint run ./... -v
	@echo "lint: all good!"
