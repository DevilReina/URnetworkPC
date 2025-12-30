#!/bin/bash
# Build script for URnetworkPC

set -e

echo "========================================="
echo "  URnetworkPC Build Script"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${BLUE}Go version: $(go version)${NC}"
echo ""

# Download dependencies
echo -e "${YELLOW}Downloading dependencies...${NC}"
go mod download
go mod tidy
echo -e "${GREEN}✓ Dependencies downloaded${NC}"
echo ""

# Build development version (with console)
echo -e "${YELLOW}Building development version (with console)...${NC}"
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
  go build -o URnetworkPC-dev.exe
echo -e "${GREEN}✓ Development build complete: URnetworkPC-dev.exe${NC}"
echo ""

# Build production version (without console, optimized)
echo -e "${YELLOW}Building production version (optimized)...${NC}"
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags="-H windowsgui -s -w" -o URnetworkPC.exe
echo -e "${GREEN}✓ Production build complete: URnetworkPC.exe${NC}"
echo ""

# Show file sizes
echo -e "${BLUE}Build artifacts:${NC}"
ls -lh URnetworkPC*.exe 2>/dev/null || ls -l URnetworkPC*.exe
echo ""

echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}  Build completed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Files created:"
echo "  • URnetworkPC-dev.exe  - Development version (with console for debugging)"
echo "  • URnetworkPC.exe      - Production version (no console, optimized)"
echo ""
echo "To run: ./URnetworkPC.exe"
