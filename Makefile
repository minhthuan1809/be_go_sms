# SMS Gateway Makefile - Optimized for Windows

# Variables
BINARY_NAME=sms-gateway
BUILD_DIR=build
MAIN_PATH=./src/cmd/server
GOFLAGS=-ldflags="-s -w"

# Quick commands for development
.PHONY: dev run build clean test hot

# Super fast development run (no flags)
dev:
	@echo "🚀 Starting development server (fast mode)..."
	go run $(MAIN_PATH)

# Fast build and run
run: build-fast start

# Hot reload with nodemon
hot:
	@echo "🔥 Starting hot reload server..."
	npm run watch

# Build commands
build-fast:
	@echo "⚡ Fast building $(BINARY_NAME)..."
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)

build:
	@echo "🔨 Building optimized $(BINARY_NAME)..."
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)

# Start the built executable
start:
	@echo "▶️ Starting $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME).exe

# Quick test
test:
	@echo "🧪 Running tests..."
	go test -v ./src/...

test-fast:
	@echo "⚡ Running tests (fast)..."
	go test ./src/...

# Clean
clean:
	@echo "🧹 Cleaning..."
	@if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod download

# Format and vet
fmt:
	@echo "✨ Formatting code..."
	go fmt ./src/...

vet:
	@echo "🔍 Vetting code..."
	go vet ./src/...

# Help
help:
	@echo "🚀 SMS Gateway - Quick Commands:"
	@echo ""
	@echo "Development:"
	@echo "  dev       - Super fast development run"
	@echo "  hot       - Hot reload with auto-restart"
	@echo "  run       - Build fast and run"
	@echo ""
	@echo "Build:"
	@echo "  build     - Optimized production build"
	@echo "  build-fast- Fast development build"
	@echo ""
	@echo "Other:"
	@echo "  test      - Run tests with output"
	@echo "  test-fast - Run tests quickly"
	@echo "  clean     - Clean build files"
	@echo "  fmt       - Format code"
