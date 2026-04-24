#!/bin/bash

# --- Sovereign Intelligence Core Launch Script ---

# Colors for terminal output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo -e "${BLUE}${BOLD}Starting Sovereign Intelligence Core Deployment...${NC}"

# Check for dependencies
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed."
    exit 1
fi

if ! command -v g++ &> /dev/null; then
    echo "Error: g++ is not installed."
    exit 1
fi

# Ensure data directory exists
mkdir -p data/sovereign

# Build the system
echo -e "${BLUE}Building components...${NC}"
make build

if [ $? -ne 0 ]; then
    echo "Build failed. Please check the errors above."
    exit 1
fi

echo -e "${GREEN}Build successful!${NC}"

# Start the Math Intelligence Core (Python)
echo -e "${BLUE}Initializing Math Intelligence Core...${NC}"
python3 math_solver/main.py &
MATH_PID=$!
echo -e "${GREEN}Math Intelligence Core started (PID: $MATH_PID)${NC}"

# Set environment variables
export SOVEREIGN_PORT=8081
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$(pwd)/internal/titan/cpp/finance

# Execute the binary
./bin/sovereign

# Kill the math service on exit
kill $MATH_PID
