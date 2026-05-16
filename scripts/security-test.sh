#!/bin/bash

# Security Testing and Verification Script for Healthcare API
# This script runs all security tests and generates reports

set -e

echo "=========================================="
echo "Healthcare API - Security Testing Suite"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check prerequisites
echo "[INFO] Checking prerequisites..."

if ! command -v go &> /dev/null; then
    echo -e "${RED}[ERROR] Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}[OK]${NC} Go is installed: $(go version)"

# Run tests
echo ""
echo "[TEST] Running unit tests..."
if go test -v ./...; then
    echo -e "${GREEN}[PASS]${NC} All unit tests passed"
else
    echo -e "${RED}[FAIL]${NC} Some tests failed"
    exit 1
fi

# Generate coverage report
echo ""
echo "[COVERAGE] Generating test coverage report..."
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
echo -e "${GREEN}[OK]${NC} Coverage report generated: coverage.html"

# Check for vulnerable dependencies
echo ""
echo "[VULN] Checking for vulnerable dependencies..."
if ! command -v govulncheck &> /dev/null; then
    echo "[INFO] Installing govulncheck..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
fi

if govulncheck ./...; then
    echo -e "${GREEN}[PASS]${NC} No vulnerable dependencies found"
else
    echo -e "${YELLOW}[WARN]${NC} Some vulnerabilities were found"
fi

# Run security scanner
echo ""
echo "[SECURITY] Running gosec security scanner..."
if ! command -v gosec &> /dev/null; then
    echo "[INFO] Installing gosec..."
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
fi

if gosec -no-fail ./...; then
    echo -e "${GREEN}[PASS]${NC} Security scan completed"
else
    echo -e "${YELLOW}[WARN]${NC} Some security issues were found"
fi

# Run go vet
echo ""
echo "[VET] Running go vet..."
if go vet ./...; then
    echo -e "${GREEN}[PASS]${NC} Go vet check passed"
else
    echo -e "${RED}[FAIL]${NC} Go vet check failed"
    exit 1
fi

# Final summary
echo ""
echo "=========================================="
echo -e "${GREEN}[SUCCESS]${NC} Security testing complete!"
echo "=========================================="
echo ""
echo "Reports generated:"
echo "  - coverage.html (test coverage)"
echo ""
echo "Next steps:"
echo "  1. Review coverage.html for test coverage"
echo "  2. Build the application: make build"
echo "  3. Run the server: make run"
echo ""
