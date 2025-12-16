.PHONY: build build-all install clean test lint fmt help

# Variables
BINARY_NAME=devops-toolkit
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Build directories
BUILD_DIR=build
DIST_DIR=dist

## help: Show this help message
help:
	@echo "DevOps Toolkit - Build Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	# macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "Built binaries in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

## install: Install to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

## install-local: Install to /usr/local/bin (requires sudo)
install-local: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	@echo "Format complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies downloaded"

## deps-update: Update dependencies
deps-update:
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "Dependencies updated"

## run: Run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "Docker image built: $(BINARY_NAME):$(VERSION)"

## release: Create a new release (requires goreleaser)
release:
	@echo "Creating release..."
	@if command -v goreleaser > /dev/null; then \
		goreleaser release --clean; \
	else \
		echo "goreleaser not installed. Install with:"; \
		echo "  go install github.com/goreleaser/goreleaser@latest"; \
	fi

## snapshot: Create a snapshot release (for testing)
snapshot:
	@echo "Creating snapshot..."
	@if command -v goreleaser > /dev/null; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "goreleaser not installed."; \
	fi

# Default target
.DEFAULT_GOAL := help

