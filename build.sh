#!/bin/bash

# Function to display usage
usage() {
    echo "Usage: $0 [--linux] [--exe] [--all]"
    echo "  --linux : Build for Linux"
    echo "  --exe   : Build for Windows (.exe)"
    echo "  --all   : Build for both Linux and Windows"
    exit 1
}

# Initialize flags
BUILD_LINUX=false
BUILD_WINDOWS=false

# Check if no arguments provided
if [ $# -eq 0 ]; then
    usage
fi

# Parse arguments
for arg in "$@"; do
    case $arg in
        --linux)
            BUILD_LINUX=true
            ;;
        --exe)
            BUILD_WINDOWS=true
            ;;
        --all)
            BUILD_LINUX=true
            BUILD_WINDOWS=true
            ;;
        *)
            echo "Unknown argument: $arg"
            usage
            ;;
    esac
done

# Create bin directory if it doesn't exist
mkdir -p bin

# Build function
build_target() {
    local os=$1
    local ext=$2
    
    echo "Building for $os..."
    
    GOOS=$os go build -o "bin/chat-server${ext}" cmd/server/main.go
    GOOS=$os go build -o "bin/chat-client${ext}" cmd/client/main.go
    GOOS=$os go build -o "bin/chat-client-tui${ext}" cmd/client-tui/main.go
    
    echo "Done building for $os."
}

# Execution based on flags
if [ "$BUILD_LINUX" = true ]; then
    build_target linux ""
fi

if [ "$BUILD_WINDOWS" = true ]; then
    build_target windows ".exe"
fi