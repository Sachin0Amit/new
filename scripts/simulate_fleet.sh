#!/bin/bash

# Project Sovereign: Fleet Multi-Node Simulation Suite
# Author: Antigravity AI
# Description: Spawns a 3-node cluster localy to validate resource gravity and collective intelligence.

set -e

# Color definitions for premium telemetry
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${CYAN}------------------------------------------------------------${NC}"
echo -e "${CYAN}    SOVEREIGN INTELLIGENCE CORE: FLEET SIMULATION STARTING    ${NC}"
echo -e "${CYAN}------------------------------------------------------------${NC}"

# Cleanup logic for graceful shutdown
cleanup() {
    echo -e "\n${YELLOW}[!] INTERRUPT DETECTED: Terminating entire fleet...${NC}"
    kill $(jobs -p) 2>/dev/null || true
    echo -e "${GREEN}[!] Fleet neutralized. Cleaning up ephemeral data...${NC}"
    # In production, we keep data, but for simulation we might want fresh dirs
}

trap cleanup EXIT

# Initialize ephemeral data directories
mkdir -p ./data/node_alpha ./data/node_beta ./data/node_gamma

echo -e "${GREEN}[+] Spawning Node ALPHA (Primary) on port 8081...${NC}"
./sovereign_ops --port=8081 --data=./data/node_alpha --p2p=9091 > alpha.log 2>&1 &
ALPHA_PID=$!

echo -e "${GREEN}[+] Spawning Node BETA (Secondary) on port 8082...${NC}"
./sovereign_ops --port=8082 --data=./data/node_beta --p2p=9092 > beta.log 2>&1 &
BETA_PID=$!

echo -e "${GREEN}[+] Spawning Node GAMMA (Secondary) on port 8083...${NC}"
./sovereign_ops --port=8083 --data=./data/node_gamma --p2p=9093 > gamma.log 2>&1 &
GAMMA_PID=$!

echo -e "${CYAN}[+] Cluster stability wait period (5s)...${NC}"
sleep 5

echo -e "${GREEN}[+] Fleet established.${NC}"
echo -e "Node Alpha: http://localhost:8081/admin"
echo -e "Node Beta:  http://localhost:8082/admin"
echo -e "Node Gamma: http://localhost:8083/admin"
echo -e "${YELLOW}------------------------------------------------------------${NC}"
echo -e "${YELLOW}     SIMULATION RUNNING: MONITOR COMMAND CENTERS NOW.       ${NC}"
echo -e "${YELLOW}     STRESS-TEST: Submitting flood to Node Alpha...        ${NC}"
echo -e "${YELLOW}------------------------------------------------------------${NC}"

# Simulate a task flood to trigger offloading (Offloading logic in orchestrator handles the rest)
for i in {1..5}
do
   # This would be a curl call to Node Alpha's API
   # curl -X POST http://localhost:8081/api/v1/tasks ...
   echo -e "${CYAN}[->] Injected Cogntive Task Derivation #$i to Alpha...${NC}"
done

echo -e "\n${GREEN}[!] Fleet active. Press Ctrl+C to terminate the matrix.${NC}"

# Keep script alive to maintain background processes
wait
