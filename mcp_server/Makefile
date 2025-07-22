# Variables
BINARY_NAME := mcp-server
CMD_DIR := ./cmd/server
VERSION ?= dev
LDFLAGS := -X github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/config.Version=$(VERSION)

# Default target
.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(CMD_DIR)

# Build with specific version
.PHONY: build-version
build-version:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required. Usage: make build-version VERSION=v1.0.0"; exit 1; fi
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-$(VERSION) $(CMD_DIR)

# Development build
.PHONY: dev
dev:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(CMD_DIR)

# Test
.PHONY: test
test:
	go test ./...

# Clean
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)*

# Run
.PHONY: run
run: build
	./$(BINARY_NAME)

# Show current version that would be built
.PHONY: version
version:
	@echo "Current version: $(VERSION)"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build with default/current VERSION ($(VERSION))"
	@echo "  build-version - Build with specific version: make build-version VERSION=v1.0.0"
	@echo "  dev           - Build development version"
	@echo "  test          - Run tests"
	@echo "  clean         - Remove built binaries"
	@echo "  run           - Build and run the server"
	@echo "  version       - Show current version"
	@echo "  help          - Show this help" 