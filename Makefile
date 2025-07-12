# Discord Gemini Bot Makefile

.PHONY: build run test test-coverage clean example help

# Default target
help:
	@echo "Available commands:"
	@echo "  build      - Build the Discord bot binary"
	@echo "  run        - Run the Discord bot"
	@echo "  test       - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  example    - Run the example usage"
	@echo "  clean      - Clean build artifacts"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  help       - Show this help message"

# Build the main Discord bot
build:
	@echo "Building Discord Gemini Bot..."
	go build -o discord-gemini-bot src/main.go

# Run the Discord bot
run: build
	@echo "Starting Discord Gemini Bot..."
	./discord-gemini-bot

# Test the project
test:
	@echo "Running tests..."
	go test -v ./tests/

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -cover ./tests/

# Run the example
example:
	@echo "Running example usage..."
	cd examples && go run example_usage.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f discord-gemini-bot

# Download and tidy dependencies
deps:
	@echo "Downloading and tidying dependencies..."
	go mod download
	go mod tidy

# Check for any issues
check:
	@echo "Running go vet..."
	go vet ./...
	@echo "Checking formatting..."
	gofmt -l .

# Install required tools
install-tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
