BINARY_NAME=reward_watcher
BUILD_DIR=build

.PHONY: all build download-abis update-abis clean test deps help

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "✅ Built $(BUILD_DIR)/$(BINARY_NAME)"

download-abis:
	@echo "Downloading ABIs from Livepeer protocol repository..."
	@go run scripts/download-abis.go
update-abis: download-abis

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

clean-all:
	@echo "Cleaning build artifacts and ABIs..."
	@rm -rf $(BUILD_DIR)
	@rm -rf ABI/

help:
	@echo "Available commands:"
	@echo "  make build         - Build the application"
	@echo "  make download-abis - Download ABIs from protocol repository"
	@echo "  make update-abis   - Update ABIs to latest versions"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make clean-all     - Clean build artifacts and ABIs"
	@echo "  make help          - Show this help"
	@echo ""
	@echo "Typical workflow:"
	@echo "  1. make download-abis  # Download ABIs (first time)"
	@echo "  2. make build          # Build the application"
	@echo "  3. make update-abis    # Update ABIs when needed"
