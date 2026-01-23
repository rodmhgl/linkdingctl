.PHONY: help build test install clean lint vet fmt cover

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

cover: ## Run tests with coverage validation (min 70% per package)
	@echo "Running tests with coverage validation..."
	@$(GO) test -cover ./... 2>&1 | tee /tmp/coverage.txt
	@echo ""
	@echo "Validating coverage thresholds (minimum 70%)..."
	@failed=0 && \
	while read -r line; do \
		if echo "$$line" | grep -q "coverage:"; then \
			pkg=$$(echo "$$line" | awk '{print $$2}'); \
			cov=$$(echo "$$line" | grep -oP '\d+\.\d+(?=% of statements)'); \
			if [ -n "$$cov" ]; then \
				result=$$(echo "$$cov 70" | awk '{print ($$1 < $$2)}'); \
				if [ "$$result" -eq 1 ]; then \
					echo "❌ FAIL: $$pkg has $$cov% coverage (below 70%)"; \
					failed=1; \
				else \
					echo "✅ PASS: $$pkg has $$cov% coverage"; \
				fi; \
			fi; \
		fi; \
	done < /tmp/coverage.txt && \
	rm -f /tmp/coverage.txt && \
	if [ $$failed -eq 1 ]; then \
		echo "" && \
		echo "Coverage validation failed. Some packages are below 70% threshold." && \
		exit 1; \
	else \
		echo "" && \
		echo "All tested packages meet the 70% coverage threshold!"; \
	fi

.DEFAULT_GOAL := help
