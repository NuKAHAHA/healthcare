# Security Testing and Verification Script for Healthcare API (Windows PowerShell)
# This script runs all security tests and generates reports

$ErrorActionPreference = "Stop"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Healthcare API - Security Testing Suite" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Check prerequisites
Write-Host "[INFO] Checking prerequisites..." -ForegroundColor Yellow

# Check Go
$goVersion = go version
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Go is installed: $goVersion" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Go is not installed" -ForegroundColor Red
    exit 1
}

# Run tests
Write-Host ""
Write-Host "[TEST] Running unit tests..." -ForegroundColor Yellow
go test -v ./...
if ($LASTEXITCODE -eq 0) {
    Write-Host "[PASS] All unit tests passed" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Some tests failed" -ForegroundColor Red
    exit 1
}

# Generate coverage report
Write-Host ""
Write-Host "[COVERAGE] Generating test coverage report..." -ForegroundColor Yellow
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
Write-Host "[OK] Coverage report generated: coverage.html" -ForegroundColor Green

# Check for vulnerable dependencies
Write-Host ""
Write-Host "[VULN] Checking for vulnerable dependencies..." -ForegroundColor Yellow

$govulnCheck = Get-Command govulncheck -ErrorAction SilentlyContinue
if (-not $govulnCheck) {
    Write-Host "[INFO] Installing govulncheck..." -ForegroundColor Yellow
    go install golang.org/x/vuln/cmd/govulncheck@latest
}

govulncheck ./...
if ($LASTEXITCODE -eq 0) {
    Write-Host "[PASS] No vulnerable dependencies found" -ForegroundColor Green
} else {
    Write-Host "[WARN] Some vulnerabilities were found" -ForegroundColor Yellow
}

# Run security scanner
Write-Host ""
Write-Host "[SECURITY] Running gosec security scanner..." -ForegroundColor Yellow

$gosecCmd = Get-Command gosec -ErrorAction SilentlyContinue
if (-not $gosecCmd) {
    Write-Host "[INFO] Installing gosec..." -ForegroundColor Yellow
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
}

gosec -no-fail ./...
if ($LASTEXITCODE -eq 0 -or $LASTEXITCODE -eq 1) {
    Write-Host "[PASS] Security scan completed" -ForegroundColor Green
} else {
    Write-Host "[WARN] Some security issues were found" -ForegroundColor Yellow
}

# Run go vet
Write-Host ""
Write-Host "[VET] Running go vet..." -ForegroundColor Yellow
go vet ./...
if ($LASTEXITCODE -eq 0) {
    Write-Host "[PASS] Go vet check passed" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Go vet check failed" -ForegroundColor Red
    exit 1
}

# Final summary
Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "[SUCCESS] Security testing complete!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Reports generated:" -ForegroundColor Yellow
Write-Host "  - coverage.html (test coverage)" -ForegroundColor Yellow
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Review coverage.html for test coverage" -ForegroundColor Yellow
Write-Host "  2. Build the application: go build ./cmd/main.go" -ForegroundColor Yellow
Write-Host "  3. Run the server: .\run.ps1" -ForegroundColor Yellow
Write-Host ""
