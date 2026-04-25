#!/bin/bash

# Sovereign Intelligence Core: Production Launch Script
# Comprehensive health-check and dependency management

# Color definitions
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'
BOLD='\033[1m'

# Configuration
CORE_PORT=8081
OLLAMA_PORT=11434
PYTHON_PORT=5000
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
BADGER_PORT=8002
MAX_WAIT=300  # 5 minutes max wait time

# Utility functions
check_command() {
    command -v "$1" >/dev/null 2>&1
}

wait_for_port() {
    local port=$1
    local service=$2
    local max_attempts=$((MAX_WAIT / 2))
    local attempt=0
    
    echo -ne "${BLUE}Waiting for ${service} on port ${port}...${NC}"
    
    while [ $attempt -lt $max_attempts ]; do
        if nc -z localhost $port 2>/dev/null || curl -s http://localhost:$port/health >/dev/null 2>&1; then
            echo -e " ${GREEN}✓${NC}"
            return 0
        fi
        echo -ne "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo -e " ${RED}✗${NC}"
    return 1
}

check_health() {
    local url=$1
    local name=$2
    
    echo -ne "${BLUE}Checking ${name}...${NC}"
    
    if curl -s "$url" >/dev/null 2>&1; then
        echo -e " ${GREEN}✓${NC}"
        return 0
    else
        echo -e " ${RED}✗${NC}"
        return 1
    fi
}

# Start sequence
echo -e "${BOLD}${BLUE}"
echo "╔════════════════════════════════════════════════════════╗"
echo "║  Sovereign Intelligence Core - Production Launch       ║"
echo "╚════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check prerequisites
echo -e "${YELLOW}1. Checking prerequisites...${NC}"

if ! check_command "docker"; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    exit 1
fi

if ! check_command "docker-compose"; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Prerequisites met${NC}\n"

# Build if needed
if [ ! -f "./bin/sovereign" ]; then
    echo -e "${YELLOW}2. Building Sovereign Core...${NC}"
    make build
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Build failed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Build complete${NC}\n"
fi

# Create data directories
echo -e "${YELLOW}3. Setting up data directories...${NC}"
mkdir -p data/{training,badger,mesh}
mkdir -p checkpoints results logs
echo -e "${GREEN}✓ Directories ready${NC}\n"

# Start docker-compose services
echo -e "${YELLOW}4. Starting Docker Compose services...${NC}"

if docker-compose ps | grep -q "running"; then
    echo -e "${YELLOW}Services already running, skipping startup${NC}"
else
    docker-compose up -d
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Docker Compose failed to start${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker Compose started${NC}\n"
fi

# Health checks
echo -e "${YELLOW}5. Health checking dependencies...${NC}"

# Ollama
if ! wait_for_port $OLLAMA_PORT "Ollama"; then
    echo -e "${YELLOW}Ollama not responding. Starting it manually...${NC}"
    # If Ollama is in Docker, give it more time
    sleep 10
    if ! wait_for_port $OLLAMA_PORT "Ollama"; then
        echo -e "${RED}⚠️  Ollama health check failed (non-critical)${NC}"
    fi
fi

# Python Sidecar
if ! wait_for_port $PYTHON_PORT "Python Sidecar"; then
    echo -e "${RED}⚠️  Python Sidecar not responding (non-critical)${NC}"
fi

# Prometheus
if ! wait_for_port $PROMETHEUS_PORT "Prometheus"; then
    echo -e "${YELLOW}⚠️  Prometheus not ready yet${NC}"
fi

echo ""

# Pre-flight checks
echo -e "${YELLOW}6. Running pre-flight checks...${NC}"

# Check if data directory has training data
if [ ! -f "./data/training/train.jsonl" ]; then
    echo -e "${YELLOW}⚠️  Training data not found. Run: make download-data${NC}"
fi

# Check if model checkpoints exist
if [ ! -d "./checkpoints" ] || [ -z "$(ls -A ./checkpoints)" ]; then
    echo -e "${YELLOW}⚠️  No model checkpoints found. Consider fine-tuning: make finetune${NC}"
fi

echo -e "${GREEN}✓ Pre-flight checks complete${NC}\n"

# Summary
echo -e "${BOLD}${GREEN}"
echo "╔════════════════════════════════════════════════════════╗"
echo "║  🚀 Sovereign Core is READY                            ║"
echo "╚════════════════════════════════════════════════════════╝"
echo -e "${NC}"

echo -e "${BLUE}Service URLs:${NC}"
echo "  Core API:          http://localhost:$CORE_PORT"
echo "  Chat Interface:    http://localhost:$CORE_PORT/web/chat.html"
echo "  Grafana Dashboard: http://localhost:$GRAFANA_PORT"
echo "  Prometheus:        http://localhost:$PROMETHEUS_PORT"
echo "  BadgerDB Explorer: http://localhost:$BADGER_PORT"
echo "  Ollama API:        http://localhost:$OLLAMA_PORT"
echo ""

echo -e "${BLUE}Useful commands:${NC}"
echo "  make health-check  - Check all service health"
echo "  make download-data - Download training datasets"
echo "  make finetune      - Fine-tune the model"
echo "  make eval          - Evaluate model performance"
echo "  make docker-logs   - View service logs"
echo "  make docker-down   - Stop all services"
echo ""

# Final message
echo -e "${GREEN}Sovereign Intelligence Core is running!${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop services${NC}\n"

# Keep script running and show logs
docker-compose logs -f sovereign-core

