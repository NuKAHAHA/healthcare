# Run the Healthcare API server
# This script builds and runs the application

$ErrorActionPreference = "Stop"

Write-Host "Starting Healthcare API server..." -ForegroundColor Cyan

# Check if Go is installed
$goVersion = go version
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go is not installed" -ForegroundColor Red
    exit 1
}

# Build the application
Write-Host "Building application..." -ForegroundColor Yellow
go build -o healthcare-api.exe ./cmd/main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Build failed" -ForegroundColor Red
    exit 1
}

Write-Host "Build successful!" -ForegroundColor Green

# Set environment variables (if needed)
$env:ENV_NAME = "development"
$env:SERVER_PORT = "8080"

# Run the application
Write-Host "Server starting on http://localhost:8080" -ForegroundColor Cyan
Write-Host ""
.\healthcare-api.exe

# Handle interruption
$null = Read-Host "Press Enter to exit"
