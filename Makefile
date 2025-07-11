# Go Image Resize Tool Makefile

BINARY_NAME=resize-tool
MAIN_PACKAGE=.
BUILD_DIR=build

all: build        ## Build the application

build:            ## Build the application
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

build-dir:        ## Build and output to build directory
	@echo "Building $(BINARY_NAME) to $(BUILD_DIR) directory..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

deps:             ## Install dependencies
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

test:             ## Run tests
	@echo "Running tests..."
	go test -v ./...

clean:            ## Clean build files
	@echo "Cleaning build files..."
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

run:              ## Show usage instructions (requires an image file argument)
	@echo "Please specify an image file: make run-example"

run-example:      ## Run example (if example image exists)
	@if [ -f "example.jpg" ]; then \
		./$(BINARY_NAME) example.jpg; \
	else \
		echo "No example.jpg found. Please provide an image file."; \
		echo "Usage: ./$(BINARY_NAME) [options] <image-file>"; \
		./$(BINARY_NAME) --help; \
	fi

release:          ## Build release versions (cross-platform compilation)
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

help:             ## Print this help message.
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

.PHONY: all build build-dir deps test clean run run-example release help
