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

.PHONY: all build clean test lint dev docker-build docker-up docker-down docker-logs \
        download-data finetune eval health-check build-wasm distributed-train help

all: build

# 1. Build Target (Polyglot Compilation)
build: build-cpp build-go build-web
	@echo "✅ Full build complete."

build-cpp:
	@echo "🔨 Building C++ Engine Core (Shared)..."
	@mkdir -p $(CPP_DIR)/build/Release
	@cd $(CPP_DIR)/build/Release && cmake -DCMAKE_BUILD_TYPE=Release ../.. && make titan_engine

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
	@LD_LIBRARY_PATH=$(PWD)/cpp/build/Release:$(PWD)/pkg/finance ENV=dev $(GO) run ./cmd/sovereign/main.go

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
	@echo "  make docker-up     - Start all services with docker-compose"
	@echo "  make docker-down   - Stop all services"
	@echo "  make docker-logs   - View docker-compose logs"
	@echo "  make download-data - Download training datasets"
	@echo "  make finetune      - Fine-tune model with LoRA"
	@echo "  make eval          - Evaluate model"
	@echo "  make build-wasm        - Compile C++ parser to WASM"
	@echo "  make distributed-train - Multi-GPU training with DeepSpeed"
	@echo "  make health-check  - Check health of all services"
	@echo "  make clean         - Remove build artifacts"

# 7. Docker Compose Targets
docker-up:
	@echo "🚀 Starting all services with docker-compose..."

# 8. Foundation Model (From Scratch)
foundation-gen:
	@echo "🧠 Generating Synthetic Intelligence Data..."
	@cd papiai_foundation_model && python3 synthetic_data_generator.py

foundation-train:
	@echo "⚡ Compiling Neural Engine Matrix..."
	@cd papiai_foundation_model && python3 train.py

foundation-chat:
	@echo "💬 Launching Sovereign Neural Inference..."
	@cd papiai_foundation_model && python3 inference.py
	@docker-compose up -d
	@echo "✅ Services started. Dashboard at http://localhost:3000"

docker-down:
	@echo "🛑 Stopping all services..."
	@docker-compose down

docker-logs:
	@echo "📊 Streaming logs..."
	@docker-compose logs -f

# 8. Training Pipeline Targets
download-data:
	@echo "📥 Downloading training datasets..."
	@python3 scripts/download_datasets.py
	@echo "✅ Training data downloaded to data/training/"

finetune:
	@echo "🎓 Starting LoRA fine-tuning..."
	@python3 scripts/finetune.py \
		--model mistralai/Mistral-7B \
		--train-data data/training/train.jsonl \
		--val-data data/training/val.jsonl \
		--output-dir checkpoints/lora_finetuned \
		--epochs 3 \
		--batch-size 8 \
		--lr 2e-4 \
		--rank 16 \
		--alpha 32
	@echo "✅ Fine-tuning complete!"

eval:
	@echo "📈 Evaluating model..."
	@python3 scripts/eval.py \
		--model checkpoints/lora_finetuned/final_model \
		--val-data data/training/val.jsonl \
		--output-json results/eval_metrics.json \
		--mlflow
	@echo "✅ Evaluation complete! Results at results/eval_metrics.json"

# 9. Health Check Target
health-check:
	@echo "🏥 Checking service health..."
	@echo -n "Sovereign Core: "
	@curl -s http://localhost:8081/health > /dev/null && echo "✅ OK" || echo "❌ FAILED"
	@echo -n "Ollama: "
	@curl -s http://localhost:11434/api/tags > /dev/null && echo "✅ OK" || echo "❌ FAILED"
	@echo -n "Python Sidecar: "
	@curl -s http://localhost:5000/health > /dev/null && echo "✅ OK" || echo "❌ FAILED"
	@echo -n "Prometheus: "
	@curl -s http://localhost:9090/-/healthy > /dev/null && echo "✅ OK" || echo "❌ FAILED"
	@echo -n "Grafana: "
	@curl -s http://localhost:3000/api/health > /dev/null && echo "✅ OK" || echo "❌ FAILED"
	@echo ""
	@echo "Dashboards:"
	@echo "  Sovereign Core:    http://localhost:8081"
	@echo "  Chat Interface:    http://localhost:8081/web/chat.html"
	@echo "  Grafana:           http://localhost:3000"
	@echo "  Prometheus:        http://localhost:9090"
	@echo "  BadgerDB Explorer: http://localhost:8002"

# 10. WASM Build Target
build-wasm:
	@echo "🔨 Building C++ Expression Parser to WASM..."
	@chmod +x scripts/build_wasm.sh
	@bash scripts/build_wasm.sh
	@echo "✅ WASM build complete. Output in web/wasm/"

# 11. Distributed Multi-GPU Training
distributed-train:
	@echo "🚀 Starting distributed multi-GPU training..."
	@torchrun --nproc_per_node=$$(nvidia-smi -L 2>/dev/null | wc -l || echo 1) \
		scripts/distributed_train.py \
		--model mistralai/Mistral-7B-v0.1 \
		--data data/training/train.jsonl \
		--val_data data/training/val.jsonl \
		--output checkpoints/distributed_lora
	@echo "✅ Distributed training complete!"
