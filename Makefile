# Makefile for openai-agents-go

.PHONY: help lint test build clean fmt

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

lint: ## Run golangci-lint
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint is not installed. Install it from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	}
	golangci-lint run ./...

fmt: ## Format code with gofmt and goimports
	gofmt -s -w .
	goimports -w -local github.com/MitulShah1/openai-agents-go .

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

coverage: test ## Show test coverage
	go tool cover -html=coverage.out

build: ## Build the project
	go build ./...

clean: ## Clean build artifacts
	rm -f coverage.out
	go clean

check: fmt lint test ## Format, lint, and test

install-tools: ## Install development tools
	@echo "Installing golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	@echo "Installing goimports..."
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Tools installed successfully!"

.DEFAULT_GOAL := help
