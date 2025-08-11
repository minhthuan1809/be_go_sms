.PHONY: build run test clean help

# Build the application
build:
	go build -o bin/sms-gateway cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run in background
run-bg:
	go run cmd/server/main.go &

# Stop background process
stop:
	pkill -f "go run cmd/server/main.go" || true

# Test the API
test:
	./test_api.sh

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Install dependencies
deps:
	go mod tidy
	go mod download

# Check for security vulnerabilities
security:
	go list -json -deps ./... | nancy sleuth

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Show help
help:
	@echo "Available commands:"
	@echo "  build    - Build the application"
	@echo "  run      - Run the application"
	@echo "  run-bg   - Run the application in background"
	@echo "  stop     - Stop background process"
	@echo "  test     - Test the API endpoints"
	@echo "  clean    - Clean build artifacts"
	@echo "  deps     - Install dependencies"
	@echo "  fmt      - Format code"
	@echo "  lint     - Lint code"
	@echo "  help     - Show this help"
