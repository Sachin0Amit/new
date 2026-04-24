# --- Sovereign Intelligence Core Build System ---
# Handles multi-language builds: Go (Orchestrator), C++ (Titan Engine), and C++ (Finance Engine)

BINARY_NAME=sovereign
BUILD_DIR=bin
CMD_PATH=./cmd/sovereign/main.go

.PHONY: all build clean test lint cores

all: cores build

# --- C++ Core Engines ---
cores:
	@echo "Building Titan Engine Core..."
	@$(MAKE) -C internal/titan/cpp
	@echo "Building Finance Engine Core..."
	@$(MAKE) -C internal/titan/cpp/finance

# --- Go Build ---
build: cores
	@echo "Building Sovereign Orchestrator..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

# --- Development Helpers ---
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

test:
	go test -v ./...

clean:
	@echo "Cleaning artifacts..."
	rm -rf $(BUILD_DIR)
	@$(MAKE) -C internal/titan/cpp clean
	@$(MAKE) -C internal/titan/cpp/finance clean
	go clean -cache

lint:
	golangci-lint run

# Setup development environment
setup:
	go mod tidy
	@mkdir -p data/sovereign
