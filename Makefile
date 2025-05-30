GOCMD=go
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOGENERATE=$(GOCMD) generate

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: all test coverage deps generate lint fmt vet check examples help

all: deps generate fmt vet test

help: ## Display this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run tests
	$(GOTEST) -v ./...

test-verbose: ## Run tests with verbose output
	$(GOTEST) -v -race ./...

test-short: ## Run short tests only
	$(GOTEST) -short ./...

coverage: ## Run tests with coverage
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean-coverage: ## Clean coverage files
	rm -f coverage.out coverage.html

coverage-func: ## Show coverage by function
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -func=coverage.out

examples: ## Run example tests
	$(GOTEST) -v ./examples/...

benchmark: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

deps-update: ## Update dependencies
	$(GOGET) -u ./...
	$(GOMOD) tidy

generate: ## Run go generate
	$(GOGENERATE) ./...

lint: ## Run golangci-lint
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

fmt: ## Format code
	$(GOCMD) fmt ./...
	goimports -w .
	golines -w .

vet: ## Run go vet
	$(GOCMD) vet ./...

check: fmt vet lint ## Run all checks (format, vet, lint)

mod-verify: ## Verify dependencies
	$(GOMOD) verify

mod-why: ## Show why a dependency is needed
	@read -p "Enter module name: " module; \
	$(GOMOD) why $$module

install-tools: ## Install development tools
	$(GOGET) github.com/mailru/easyjson/easyjson@latest
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

watch-test: ## Watch for changes and run tests
	@which inotifywait > /dev/null || (echo "inotifywait not found. Install inotify-tools package" && exit 1)
	@echo "Watching for changes..."
	@while true; do \
		inotifywait -r -e modify,create,delete --include='.*\.go$$' . 2>/dev/null; \
		echo "Running tests..."; \
		$(MAKE) test-short; \
		echo "Waiting for changes..."; \
	done

tag: ## Create a new tag (usage: make tag VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required. Usage: make tag VERSION=v1.0.0"; exit 1; fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

git-hooks: ## Install git hooks
	@echo "Installing git hooks..."
	@echo "#!/bin/sh\nmake check" > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed"

ci: deps generate check test coverage ## Run all CI checks
