.PHONY: build clean install help

# Build directory
BUILD_DIR := bin

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOINSTALL := $(GOCMD) install
BINARY_NAME := pw-
BINARY_UNIX := $(BINARY_NAME)_unix

# CLI tools
CLI_TOOLS := pw-list pw-monitor pw-info pw-connect

## help: Display this help message
help:
	@echo "PipeWire-Go CLI Tools Build"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build all CLI tools
build: $(patsubst %,build-%,$(CLI_TOOLS))

build-%:
	@echo "Building $*..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$* ./cmd/$*

## install: Install all CLI tools
install: build
	@echo "Installing CLI tools..."
	@$(foreach tool,$(CLI_TOOLS), \
		install -m 755 $(BUILD_DIR)/$(tool) $(GOPATH)/bin/; \
	)
	@echo "Installed to $(GOPATH)/bin/"

## clean: Remove built binaries
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

## test: Run tests
test:
	@echo "Running tests..."
	$(GOCMD) test -v -race ./...

## test-integration: Run integration tests with Docker
test-integration:
	@echo "Starting PipeWire services..."
	@docker-compose -f tests/docker-compose.yml up -d
	@sleep 5
	@echo "Running integration tests..."
	$(GOCMD) test -v -tags=integration ./tests/... || true
	@echo "Stopping services..."
	@docker-compose -f tests/docker-compose.yml down

## docs: Generate documentation
docs:
	@echo "Generating documentation..."
	$(GOCMD) doc ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@$(GOCMD) fmt ./...
	@goimports -w .

## lint: Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./...

all: clean build
