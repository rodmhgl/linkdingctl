.PHONY: help build test install clean lint vet fmt

# Variables
BINARY_NAME=ld
BUILD_DIR=.
INSTALL_DIR=/usr/local/bin
GO=go
GOFLAGS=-trimpath
LDFLAGS=-s -w

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ld

test: ## Run tests
	$(GO) test -v ./...

test-coverage: ## Run tests with coverage
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

install: build ## Install the binary to $(INSTALL_DIR)
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@echo "Installed successfully!"

uninstall: ## Remove the binary from $(INSTALL_DIR)
	@echo "Removing $(BINARY_NAME) from $(INSTALL_DIR)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstalled successfully!"

clean: ## Remove build artifacts
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

vet: ## Run go vet
	$(GO) vet ./...

fmt: ## Format code with gofmt
	$(GO) fmt ./...

check: fmt vet test ## Run format, vet, and tests

.DEFAULT_GOAL := help
