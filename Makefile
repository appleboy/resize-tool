# Go Image Resize Tool Makefile

# 變數定義
BINARY_NAME=resize-tool
MAIN_PACKAGE=.
BUILD_DIR=build

# 預設目標
all: build

# 建構應用程式
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

# 建構並輸出到 build 目錄
build-dir:
	@echo "Building $(BINARY_NAME) to $(BUILD_DIR) directory..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# 安裝依賴
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# 執行測試
test:
	@echo "Running tests..."
	go test -v ./...

# 清理建構檔案
clean:
	@echo "Cleaning build files..."
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

# 執行程式（需要提供圖片檔案參數）
run:
	@echo "Please specify an image file: make run-example"

# 執行範例（如果有測試圖片）
run-example:
	@if [ -f "example.jpg" ]; then \
		./$(BINARY_NAME) example.jpg; \
	else \
		echo "No example.jpg found. Please provide an image file."; \
		echo "Usage: ./$(BINARY_NAME) [options] <image-file>"; \
		./$(BINARY_NAME) --help; \
	fi

# 建立發布版本（跨平台編譯）
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

# 顯示幫助
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

# 定義 phony 目標
.PHONY: all build build-dir deps test clean run run-example release help
