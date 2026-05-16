# Test script for Healthcare API
# Runs comprehensive test suite with coverage and reports

$ErrorActionPreference = "Stop"

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "Healthcare API - Test Suite" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Parse command line arguments
$verbose = $args -contains "-v" -or $args -contains "--verbose"
$coverage = $args -contains "-c" -or $args -contains "--coverage"
$short = $args -contains "-s" -or $args -contains "--short"

# Run all tests
if ($short) {
    Write-Host "[TEST] Running short test suite..." -ForegroundColor Yellow
    go test ./...
} else {
    Write-Host "[TEST] Running comprehensive test suite..." -ForegroundColor Yellow
    $testArgs = @("-v", "-race", "-count=1")
    if ($coverage) {
        $testArgs += "-coverprofile=coverage.out"
    }
    go test @testArgs ./...
}

if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Tests failed!" -ForegroundColor Red
    exit 1
}

Write-Host "[PASS] All tests passed!" -ForegroundColor Green

# Generate coverage report if requested
if ($coverage -or $args -contains "--html-coverage") {
    Write-Host ""
    Write-Host "[COVERAGE] Generating coverage report..." -ForegroundColor Yellow
    
    if (-not (Test-Path coverage.out)) {
        go test -coverprofile=coverage.out ./...
    }
    
    go tool cover -html=coverage.out -o coverage.html
    Write-Host "[OK] Coverage report: coverage.html" -ForegroundColor Green
    
    # Display coverage stats
    $coverageStats = go tool cover -func=coverage.out | Select-Object -Last 1
    Write-Host "Coverage: $coverageStats" -ForegroundColor Gray
}

# Success
Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "[SUCCESS] Testing complete!" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Usage:" -ForegroundColor Yellow
Write-Host "  .\test.ps1              # Run tests" -ForegroundColor Yellow
Write-Host "  .\test.ps1 -c           # Run with coverage" -ForegroundColor Yellow
Write-Host "  .\test.ps1 -v           # Run in verbose mode" -ForegroundColor Yellow
Write-Host "  .\test.ps1 -s           # Run short suite" -ForegroundColor Yellow
Write-Host ""
