# --- Sovereign Intelligence Core Build System ---
# Handles multi-language builds: Go (Orchestrator) and C++ (Titan Engine)

BINARY_NAME=sovereign
BUILD_DIR=bin
CMD_PATH=./cmd/sovereign/main.go

.PHONY: all build clean test lint wasm-graphics

all: build wasm-graphics

# --- Go Build ---
build:
	@echo "Building Sovereign Orchestrator..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

# --- C++ WASM Build (Requires Emscripten) ---
wasm-graphics:
	@echo "Compiling Biomechanical Graphics Engine to WASM..."
	@mkdir -p public/wasm
	emcc internal/titan/graphics.cpp -O3 \
		-s WASM=1 \
		-s EXPORTED_FUNCTIONS='["_generate_ribs", "_generate_pipes"]' \
		-s EXPORTED_RUNTIME_METHODS='["ccall", "cwrap"]' \
		-o public/wasm/titan_geo.js

# --- Development Helpers ---
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR) public/wasm/titan_geo.*
	go clean -cache

lint:
	golangci-lint run
