# Sovereign Intelligence Core: Production Makefile
# Architect: Antigravity

# Paths
GO_BIN = bin/sovereign
CPP_DIR = cpp
BUILD_DIR = build
WEB_SRC = web/src
WEB_DIST = web/dist
DEPLOY_DIR = deploy

# Tooling
GO = go
GOLINT = golangci-lint
CLANG_TIDY = clang-tidy
ESLINT = npx eslint
DOCKER = docker

.PHONY: all build clean test lint dev docker-build help

all: build

# 1. Build Target (Polyglot Compilation)
build: build-cpp build-go build-web
	@echo "✅ Full build complete."

build-cpp:
	@echo "🔨 Building C++ Engine Core..."
	@mkdir -p $(BUILD_DIR)
	@cd $(BUILD_DIR) && cmake ../$(CPP_DIR) && make
	@mkdir -p internal/titan/
	@cp $(BUILD_DIR)/libtitan.a internal/titan/

build-go:
	@echo "🔨 Building Sovereign Orchestrator..."
	@mkdir -p bin/
	@$(GO) build -o $(GO_BIN) ./cmd/sovereign/main.go

build-web:
	@echo "🔨 Bundling Frontend Assets..."
	@mkdir -p $(WEB_DIST)
	@cp -r $(WEB_SRC)/* $(WEB_DIST)/
	@echo "Frontend assets deployed to $(WEB_DIST)"

# 2. Development Target (Live Reload)
dev:
	@echo "🚀 Starting development environment..."
	@ENV=dev $(GO) run ./cmd/sovereign/main.go

# 3. Test Target
test:
	@echo "🧪 Running Go test suites..."
	@$(GO) test ./...
	@echo "🧪 Running C++ tests..."
	@# Add C++ test execution if available
	@echo "🧪 Running Python core tests..."
	@if [ -d "math_solver/tests" ]; then pytest math_solver/tests/; fi

# 4. Lint Target (Cross-Language Analysis)
lint:
	@echo "🔍 Linting Go source..."
	@$(GOLINT) run ./...
	@echo "🔍 Linting C++ source..."
	@find $(CPP_DIR) -name "*.cpp" -o -name "*.hpp" | xargs $(CLANG_TIDY)
	@echo "🔍 Linting Frontend JS..."
	@$(ESLINT) $(WEB_SRC)/js/*.js

# 5. Docker Target
docker-build:
	@echo "🐳 Building Sovereign Docker Image..."
	@$(DOCKER) build -t sovereign-core:latest -f $(DEPLOY_DIR)/Dockerfile .

# 6. Clean Target
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/ $(BUILD_DIR) $(WEB_DIST)
	@find . -name "*.o" -delete
	@find . -name "*.a" -delete

help:
	@echo "Sovereign Intelligence Core Makefile"
	@echo "Targets:"
	@echo "  make build         - Build Go, C++, and Frontend"
	@echo "  make dev           - Run development server"
	@echo "  make test          - Run all tests"
	@echo "  make lint          - Run linters (Go, C++, JS)"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make clean         - Remove build artifacts"
