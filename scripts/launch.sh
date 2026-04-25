#!/bin/bash

# Sovereign Intelligence Core: Professional Launch Script
# Port: 8081

# Color definitions
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' 
BOLD='\033[1m'

# Build components if needed
if [ ! -f "./bin/sovereign" ]; then
    echo -e "${YELLOW}Binary not found. Triggering build...${NC}"
    make build
fi

# Start the Math Intelligence Core (Python)
echo -e "${BLUE}Initializing Math Intelligence Core...${NC}"
python3 math_solver/main.py &
MATH_PID=$!
echo -e "${GREEN}Math Intelligence Core started (PID: $MATH_PID)${NC}"

# Set environment variables
export SOVEREIGN_PORT=8081
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$(pwd)/build

# Execute the binary
echo -e "${BLUE}Launching Sovereign Core...${NC}"
./bin/sovereign

# Cleanup on exit
kill $MATH_PID
