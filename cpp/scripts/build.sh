#!/bin/bash
set -e

# Detect OS
OS="$(uname -s)"
echo "Detected OS: $OS"

# Install Dependencies (Simplified example for common systems)
case "$OS" in
    Linux)
        if command -v apt-get &> /dev/null; then
            echo "Using apt-get to install dependencies..."
            # sudo apt-get update && sudo apt-get install -y cmake build-essential libomp-dev
        fi
        ;;
    Darwin)
        if command -v brew &> /dev/null; then
            echo "Using Homebrew to install dependencies..."
            # brew install cmake libomp
        fi
        ;;
    *)
        echo "Unsupported OS: $OS. Please install CMake and OpenMP manually."
        ;;
esac

# Build Release
echo "Building Release..."
mkdir -p build-release
cd build-release
cmake -DCMAKE_BUILD_TYPE=Release ..
cmake --build . -j$(nproc 2>/dev/null || sysctl -n hw.ncpu)
cd ..

# Build Debug with Sanitizers
echo "Building Debug with Sanitizers..."
mkdir -p build-debug
cd build-debug
cmake -DCMAKE_BUILD_TYPE=Debug ..
cmake --build . -j$(nproc 2>/dev/null || sysctl -n hw.ncpu)

# Run Tests
echo "Running Tests..."
ctest --output-on-failure
cd ..

echo "Titan Engine Build Complete."
