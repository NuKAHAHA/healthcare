# Build script for Healthcare API
# Builds the application and generates artifacts

$ErrorActionPreference = "Stop"

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "Healthcare API - Build Script" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
Write-Host "[CHECK] Verifying Go installation..." -ForegroundColor Yellow
$goVersion = go version
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go is not installed" -ForegroundColor Red
    exit 1
}
Write-Host "OK: $goVersion" -ForegroundColor Green

# Format code
Write-Host ""
Write-Host "[FORMAT] Formatting code..." -ForegroundColor Yellow
go fmt ./...
Write-Host "OK: Code formatted" -ForegroundColor Green

# Download dependencies
Write-Host ""
Write-Host "[DEPS] Downloading dependencies..." -ForegroundColor Yellow
go mod download
Write-Host "OK: Dependencies downloaded" -ForegroundColor Green

# Verify dependencies
Write-Host ""
Write-Host "[VERIFY] Verifying dependencies..." -ForegroundColor Yellow
go mod verify
Write-Host "OK: Dependencies verified" -ForegroundColor Green

# Run tests
Write-Host ""
Write-Host "[TEST] Running tests..." -ForegroundColor Yellow
go test -v ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Tests failed" -ForegroundColor Red
    exit 1
}
Write-Host "OK: All tests passed" -ForegroundColor Green

# Build binary
Write-Host ""
Write-Host "[BUILD] Building application..." -ForegroundColor Yellow

$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$version = git describe --tags --always 2>$null
if ($LASTEXITCODE -ne 0) {
    $version = "dev"
}

go build `
    -ldflags "-X 'main.version=$version' -X 'main.buildTime=$timestamp'" `
    -o healthcare-api.exe `
    ./cmd/main.go

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Build failed" -ForegroundColor Red
    exit 1
}
Write-Host "OK: Binary created (healthcare-api.exe)" -ForegroundColor Green

# Get binary info
$fileInfo = Get-Item healthcare-api.exe
Write-Host "    Size: $($fileInfo.Length) bytes" -ForegroundColor Gray
Write-Host "    Version: $version" -ForegroundColor Gray

# Success
Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "[SUCCESS] Build completed!" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Run: .\scripts\run.ps1" -ForegroundColor Yellow
Write-Host "  2. Or: .\healthcare-api.exe" -ForegroundColor Yellow
Write-Host ""
