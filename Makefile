# Load environment variables
ifneq (,$(wildcard .env))
    include .env
    export
endif

# Default values
SERVER_URL ?= http://localhost:8080
# CLIENT_SECRET must be provided - no default for security
ifndef CLIENT_SECRET
$(error CLIENT_SECRET environment variable is required. Please set it in your .env file or environment)
endif

# Build flags
BUILD_FLAGS = -ldflags="-s -w"
SECURE_BUILD_FLAGS = -ldflags="-s -w -X 'main.BuildServerURL=$(SERVER_URL)' -X 'main.BuildClientSecret=$(CLIENT_SECRET)'"

# Development build (requires .env file)
build-dev:
	@echo "Building development version..."
	@if [ ! -f .env ]; then cp .env.dev .env; fi
	go build $(BUILD_FLAGS) -o planning_poker_dev.exe ./cmd/client
	go build $(BUILD_FLAGS) -o server.exe ./cmd/server

# Production build with embedded config
build-secure:
	@echo "Building secure version with embedded config..."
	@if [ -z "$(CLIENT_SECRET)" ]; then echo "Error: CLIENT_SECRET not set"; exit 1; fi
	go build $(SECURE_BUILD_FLAGS) -o planning_poker_secure.exe ./cmd/client

# Build all platforms (requires CLIENT_SECRET)
build-all:
	@echo "Building for all platforms..."
	@if [ -z "$(CLIENT_SECRET)" ]; then echo "Error: CLIENT_SECRET not set"; exit 1; fi
	GOOS=windows GOARCH=amd64 go build $(SECURE_BUILD_FLAGS) -o planning_poker_secure_windows_amd64.exe ./cmd/client
	GOOS=linux GOARCH=amd64 go build $(SECURE_BUILD_FLAGS) -o planning_poker_secure_linux_amd64 ./cmd/client
	GOOS=linux GOARCH=arm64 go build $(SECURE_BUILD_FLAGS) -o planning_poker_secure_linux_arm64 ./cmd/client
	GOOS=darwin GOARCH=amd64 go build $(SECURE_BUILD_FLAGS) -o planning_poker_secure_darwin_amd64 ./cmd/client
	GOOS=darwin GOARCH=arm64 go build $(SECURE_BUILD_FLAGS) -o planning_poker_secure_darwin_arm64 ./cmd/client

# Legacy build
build:
	@echo "Building legacy version..."
	go build -o server.exe ./cmd/server
	go build -o cli_planning_poker.exe ./cmd/client

# Run development server
run-server:
	@if [ ! -f .env ]; then cp .env.dev .env; fi
	go run ./cmd/server

# Run development client
run-client:
	@if [ ! -f .env ]; then cp .env.dev .env; fi
	go run ./cmd/client

# Clean build artifacts
clean:
	del /Q /F *.exe planning_poker_secure_* 2>nul || true
	rm -f planning_poker_secure_* 2>/dev/null || true

# Setup development environment
setup-dev:
	@if [ ! -f .env ]; then cp .env.dev .env; echo "Created .env from .env.dev"; fi
	go mod tidy
	@echo "Development environment ready!"
	@echo "Edit .env file to customize configuration"

.PHONY: build build-dev build-secure build-all run-server run-client clean setup-dev