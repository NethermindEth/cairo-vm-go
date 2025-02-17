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
		if [ ! -d rust_vm_bin/ctj/ctj ]; then \
			mkdir -p rust_vm_bin/ctj/ctj; \
		fi; \
		if [ ! -d rust_vm_bin/starkware/starkware ]; then \
			mkdir -p rust_vm_bin/starkware/starkware; \
		fi; \
		if [ ! -d rust_vm_bin/lambdaclass/lambdaclass ]; then \
			mkdir -p rust_vm_bin/lambdaclass/lambdaclass; \
		fi; \
		if [ ! -f ./rust_vm_bin/ctj/ctj/cairo-to-json ]; then \
			cd rust_vm_bin/ctj/ctj && \
			git clone --single-branch --branch cairo-to-json --depth=1 https://github.com/MaksymMalicki/cairo-json.git && \
			cd cairo-json/crates/bin && cargo build --release --bin cairo-to-json && \
			cd ../../../ && \
			mv cairo-json/target/release/cairo-to-json . && \
			mv cairo-json/corelib ../ && \
			rm -rf cairo-json && \
			cd ../../../; \
		fi; \
		if [ ! -f ./rust_vm_bin/starkware/starkware/cairo-run ]; then \
			cd rust_vm_bin/starkware/starkware && \
			git clone https://github.com/starkware-libs/cairo.git && \
			mv cairo/corelib ../ && \
			cd cairo/crates/bin && cargo build --release --bin cairo-run && \
			cd ../../../ && \
			mv cairo/target/release/cairo-run . && \
			rm -rf cairo && \
			cd ../../../; \
		fi; \
		if [ ! -f ./rust_vm_bin/lambdaclass/lambdaclass/cairo1-run ] || [ ! -f ./rust_vm_bin/lambdaclass/lambdaclass/cairo-vm-cli ]; then \
			cd rust_vm_bin/lambdaclass/lambdaclass && \
			git clone https://github.com/lambdaclass/cairo-vm.git && \
			cd cairo-vm/cairo1-run && make deps && \
			cd ../../cairo-vm && cargo build --release --bin cairo-vm-cli --bin cairo1-run && \
			cd ../ && \
			mv cairo-vm/target/release/cairo1-run cairo-vm/target/release/cairo-vm-cli . && \
			mv cairo-vm/cairo1-run/corelib ../ && \
			rm -rf cairo-vm && \
			cd ../../../; \
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
