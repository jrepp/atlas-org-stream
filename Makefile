# Atlassian Organization Tool Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Binary names
BINARY_NAME=atlassian-org-tool
BINARY_UNIX=$(BINARY_NAME)-linux
BINARY_WINDOWS=$(BINARY_NAME)-windows.exe
BINARY_DARWIN=$(BINARY_NAME)-darwin

# Build output directory
BUILD_DIR=build

# Version and build info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Linker flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

.PHONY: all build clean test test-coverage test-race deps lint fmt vet help
.PHONY: build-linux build-windows build-darwin build-all
.PHONY: install uninstall run

# Default target
all: clean deps test build

## Build targets
build: ## Build the binary for current platform
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -v .

build-linux: ## Build the binary for Linux
	@echo "Building $(BINARY_UNIX)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) -v .

build-windows: ## Build the binary for Windows
	@echo "Building $(BINARY_WINDOWS)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_WINDOWS) -v .

build-darwin: ## Build the binary for macOS
	@echo "Building $(BINARY_DARWIN)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_DARWIN) -v .

build-all: build-linux build-windows build-darwin ## Build binaries for all platforms

## Development targets
run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) --help

install: ## Install the binary to $GOPATH/bin
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) .

uninstall: ## Remove the binary from $GOPATH/bin
	rm -f $(GOPATH)/bin/$(BINARY_NAME)

## Testing targets
test: ## Run unit tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	$(GOTEST) -v -race ./...

test-bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	$(GOTEST) -v -bench=. -benchmem ./...

## Code quality targets
fmt: ## Format Go code
	@echo "Formatting code..."
	$(GOFMT) -s -w .

fmt-check: ## Check if code is formatted
	@echo "Checking code formatting..."
	@test -z "$(shell $(GOFMT) -s -l . | tee /dev/stderr)"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

lint-install: ## Install golangci-lint
	@echo "Installing golangci-lint..."
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

staticcheck: ## Run staticcheck
	@echo "Running staticcheck..."
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

staticcheck-install: ## Install staticcheck
	@echo "Installing staticcheck..."
	$(GOCMD) install honnef.co/go/tools/cmd/staticcheck@latest

## Security targets
security: ## Run security scan with gosec
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

security-install: ## Install gosec
	@echo "Installing gosec..."
	$(GOCMD) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

## Dependency management
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	$(GOMOD) verify

deps-tidy: ## Tidy up dependencies
	@echo "Tidying up dependencies..."
	$(GOMOD) tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

## Cleanup targets
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

clean-all: clean ## Clean everything including dependencies cache
	$(GOCLEAN) -modcache

## CI/CD targets
ci: deps fmt-check vet test-race lint staticcheck security ## Run all CI checks

## Release targets
release-check: ## Check if ready for release
	@echo "Checking release readiness..."
	@if [ -z "$(VERSION)" ]; then echo "VERSION not set"; exit 1; fi
	@git diff --quiet || (echo "Working directory is dirty"; exit 1)
	@echo "Ready for release $(VERSION)"

## Utility targets
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

size: ## Show binary size
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		echo "Binary size:"; \
		ls -lh $(BUILD_DIR)/$(BINARY_NAME) | awk '{print $$5 "\t" $$9}'; \
	else \
		echo "Binary not found. Run 'make build' first."; \
	fi

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container..."
	docker run --rm -it $(BINARY_NAME):$(VERSION) --help

## Help target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)