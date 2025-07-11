# Go Image Resize Tool Makefile

# Variable definitions
BINARY_NAME=resize-tool
MAIN_PACKAGE=.
BUILD_DIR=build

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

# Build and output to build directory
build-dir:
	@echo "Building $(BINARY_NAME) to $(BUILD_DIR) directory..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build files
clean:
	@echo "Cleaning build files..."
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

# Run the program (requires an image file argument)
run:
	@echo "Please specify an image file: make run-example"

# Run example (if example image exists)
run-example:
	@if [ -f "example.jpg" ]; then \
		./$(BINARY_NAME) example.jpg; \
	else \
		echo "No example.jpg found. Please provide an image file."; \
		echo "Usage: ./$(BINARY_NAME) [options] <image-file>"; \
		./$(BINARY_NAME) --help; \
	fi

# Build release versions (cross-platform compilation)
release:
	@echo "Building release versions..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	
	# macOS amd64
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	
	# macOS arm64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	
	@echo "Release builds completed in $(BUILD_DIR)/"

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  build-dir  - Build to build/ directory"
	@echo "  deps       - Install dependencies"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build files"
	@echo "  run        - Show usage instructions"
	@echo "  release    - Build for multiple platforms"
	@echo "  help       - Show this help"

# Define phony targets
.PHONY: all build build-dir deps test clean run run-example release help
