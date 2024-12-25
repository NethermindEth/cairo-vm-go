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
		if [ ! -d rust_vm_bin ]; then \
			mkdir -p rust_vm_bin; \
		fi; \
		if [ ! -d rust_vm_bin/cairo ]; then \
			mkdir -p rust_vm_bin/cairo-lang; \
		fi; \
		if [ ! -f rust_vm_bin/cairo/cairo-compile ] || [ ! -f rust_vm_bin/cairo/sierra-compile-json ] || [ ! -d rust_vm_bin/corelib ]; then \
			cd rust_vm_bin; \
			git clone --single-branch --branch feat/main-casm-json --depth=1 https://github.com/zmalatrax/cairo.git; \
			mv cairo/corelib .; \
			cd cairo/crates/bin && cargo build --release --bin cairo-compile --bin sierra-compile-json && cd ../../../; \
			mv cairo/target/release/cairo-compile cairo/target/release/sierra-compile-json cairo-lang; \
			rm -rf cairo; \
			cd ../; \
		fi; \
		if [ ! -f ./rust_vm_bin/cairo/cairo1-run ] || [ ! -f ./rust_vm_bin/cairo-vm-cli ]; then \
			cd rust_vm_bin; \
			git clone https://github.com/lambdaclass/cairo-vm.git; \
			cd cairo-vm && cargo build --release --bin cairo-vm-cli --bin cairo1-run && cd ../; \
			mv cairo-vm/target/release/cairo1-run cairo;\
			mv cairo-vm/target/release/cairo-vm-cli . ; \
			rm -rf cairo-vm; \
			cd ../; \
		fi; \
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

zerobench:
	@echo "Running integration benchmarks..."
	@go test integration_tests/cairozero_test.go -v -zerobench;


# Use the same version of the golangci-lint as in our CI linting config.
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3
	golangci-lint run ./... -v
	@echo "lint: all good!"
