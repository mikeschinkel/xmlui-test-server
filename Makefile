# Makefile for xmlui-test-server
#
# This Makefile provides a consistent interface for building the server
# across different platforms. Complex build logic is delegated to shell scripts.

.DEFAULT_GOAL := help

# Import variables from scripts/vars.sh (single source of truth)
BIN_DIR := $(shell bash -c 'source scripts/vars.sh && echo $$BIN_DIR')
BINARY_NAME := $(shell bash -c 'source scripts/vars.sh && echo $$BINARY_NAME')
BINARY_PATH := $(shell bash -c 'source scripts/vars.sh && echo $$BINARY_PATH')
STEAMPIPE_EXTENSION := $(shell bash -c 'source scripts/vars.sh && echo $$STEAMPIPE_EXTENSION')

## help: Show this help message
.PHONY: help
help:
	@echo "xmlui-test-server Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build standard binary without extension support
.PHONY: build
build: clean
	@./scripts/build.sh

## build-ext-macos: Build with extension support for macOS (auto-detects architecture, requires patched go-sqlite3)
.PHONY: build-ext-macos
build-ext-macos: clean
	@./scripts/build-ext.sh macos

## build-ext-macos-arm: Build with extension support for macOS ARM (Apple Silicon, requires patched go-sqlite3)
.PHONY: build-ext-macos-arm
build-ext-macos-arm: clean
	@./scripts/build-ext.sh macos arm

## build-ext-macos-intel: Build with extension support for macOS Intel (x86_64, requires patched go-sqlite3)
.PHONY: build-ext-macos-intel
build-ext-macos-intel: clean
	@./scripts/build-ext.sh macos intel

## build-ext-linux: Build with extension support for Linux AMD64
.PHONY: build-ext-linux
build-ext-linux: clean
	@./scripts/build-ext.sh linux

## install-mac: Download and install prebuilt macOS ARM binary
.PHONY: install-mac
install-mac:
	@./scripts/install.sh macos

## run: Run the server (basic mode)
.PHONY: run
run: build
	./$(BINARY_PATH)

## run-ext: Run the server with extension loading (if available)
.PHONY: run-ext
run-ext:
	@if [ -f $(STEAMPIPE_EXTENSION) ]; then \
		./$(BINARY_PATH) --extension ./$(STEAMPIPE_EXTENSION); \
	else \
		echo "Extension not found. Run 'make build-ext-macos' or 'make build-ext-linux' first."; \
		exit 1; \
	fi

## test: Run tests
.PHONY: test
test:
	@go test -v \
		./xmluisvr/... \
		./cmd/... \
	  	./test/...

## clean: Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	rm -f ext.tar.gz
	rm -f $(STEAMPIPE_EXTENSION)
	rm -rf xmlui-test-server-build

## deps: Download Go dependencies
.PHONY: deps
deps:
	cd xmluisvr && go mod download && go mod tidy
	cd cmd && go mod download && go mod tidy