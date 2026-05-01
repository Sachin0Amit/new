#!/bin/bash

# Sovereign Intelligence Core - Startup Script
# May 1, 2026

set -e

BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║   SOVEREIGN INTELLIGENCE CORE - STARTUP SCRIPT             ║"
echo "║   Version 1.2.0 - Production Grade                        ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Change to project directory
cd "$(dirname "$0")"
PROJECT_DIR=$(pwd)

echo -e "${YELLOW}[1/6] Checking prerequisites...${NC}"
if ! command -v docker &> /dev/null; then
    echo -e "${RED}ERROR: Docker is not installed${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}ERROR: Docker Compose is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Docker and Docker Compose found${NC}"

echo -e "${YELLOW}[2/6] Cleaning up old containers...${NC}"
docker-compose -f docker-compose.simple.yml down 2>/dev/null || true
docker system prune -f > /dev/null 2>&1 || true
echo -e "${GREEN}✓ Cleanup complete${NC}"

echo -e "${YELLOW}[3/6] Starting Docker services...${NC}"
docker-compose -f docker-compose.simple.yml up -d
echo -e "${GREEN}✓ Docker services starting...${NC}"

echo -e "${YELLOW}[4/6] Waiting for services to be ready (60 seconds)...${NC}"
sleep 60

echo -e "${YELLOW}[5/6] Verifying services...${NC}"

# Check Ollama
if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Ollama LLM running${NC}"
else
    echo -e "${YELLOW}⚠ Ollama is starting... (may take longer)${NC}"
fi

# Check Prometheus
if curl -s http://localhost:9090/-/healthy > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Prometheus running${NC}"
else
    echo -e "${YELLOW}⚠ Prometheus is starting...${NC}"
fi

# Check Grafana
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Grafana running${NC}"
else
    echo -e "${YELLOW}⚠ Grafana is starting...${NC}"
fi

echo -e "${YELLOW}[6/6] System ready!${NC}"

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                    🎉 READY TO USE!                       ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"

echo ""
echo -e "${BLUE}📊 ACCESS POINTS:${NC}"
echo -e "   Dashboard:        ${BLUE}http://localhost:8081/web/index-pro.html${NC}"
echo -e "   Chat Interface:   ${BLUE}http://localhost:8081/web/chat.html${NC}"
echo -e "   Finance Pro:      ${BLUE}http://localhost:8081/web/src/finance-pro.html${NC}"
echo -e "   Command Center:   ${BLUE}http://localhost:8081/web/command-center.html${NC}"
echo -e "   Status Check:     ${BLUE}http://localhost:8081/web/status.html${NC}"

echo ""
echo -e "${BLUE}🔧 SERVICE PORTS:${NC}"
echo -e "   Ollama LLM:       ${BLUE}http://localhost:11434${NC}"
echo -e "   Prometheus:       ${BLUE}http://localhost:9090${NC}"
echo -e "   Grafana:          ${BLUE}http://localhost:3000${NC} (admin / sovereign)"

echo ""
echo -e "${BLUE}📝 USEFUL COMMANDS:${NC}"
echo -e "   View logs:        ${BLUE}docker-compose -f docker-compose.simple.yml logs -f${NC}"
echo -e "   Stop services:    ${BLUE}docker-compose -f docker-compose.simple.yml down${NC}"
echo -e "   Check status:     ${BLUE}docker-compose -f docker-compose.simple.yml ps${NC}"

echo ""
echo -e "${YELLOW}💡 NEXT STEPS:${NC}"
echo -e "   1. Open Dashboard in your browser"
echo -e "   2. Click on Chat to start talking to AI"
echo -e "   3. Explore Finance Pro for trading analysis"
echo -e "   4. Monitor system via Command Center"
echo -e "   5. Check health via Status page"

echo ""
echo -e "${GREEN}✅ All systems operational!${NC}"
echo ""

# Keep script running and show service status
echo -e "${BLUE}Service Status:${NC}"
docker-compose -f docker-compose.simple.yml ps

echo ""
echo -e "${YELLOW}Press Ctrl+C to stop monitoring${NC}"
echo -e "${BLUE}Running: docker-compose -f docker-compose.simple.yml logs -f${NC}"
echo ""

# Follow logs
docker-compose -f docker-compose.simple.yml logs -f
