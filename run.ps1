# Healthcare project — start/stop helper
# Usage:
#   .\run.ps1          — start everything
#   .\run.ps1 stop     — stop everything

param([string]$Action = "start")

$root     = $PSScriptRoot
$frontend = Join-Path $root "frontend"
$binary   = Join-Path $root "backend.exe"

if ($Action -eq "stop") {
    Write-Host "Stopping backend..."
    Stop-Process -Name "backend" -ErrorAction SilentlyContinue

    Write-Host "Stopping frontend (node)..."
    Stop-Process -Name "node" -ErrorAction SilentlyContinue

    Write-Host "All stopped."
    exit 0
}

# ── BUILD backend if binary is missing or source is newer ──────────
$mainGo = Join-Path $root "cmd\main.go"
$rebuild = (-not (Test-Path $binary)) -or
           ((Get-Item $mainGo).LastWriteTime -gt (Get-Item $binary).LastWriteTime)

if ($rebuild) {
    Write-Host "Building backend..."
    & go build -o $binary (Join-Path $root "cmd\main.go")
    if ($LASTEXITCODE -ne 0) { Write-Error "Build failed"; exit 1 }
}

# ── START backend ──────────────────────────────────────────────────
Write-Host "Starting backend  -> http://localhost:8081"
Start-Process -FilePath $binary -WorkingDirectory $root -WindowStyle Hidden

# ── START frontend ─────────────────────────────────────────────────
Write-Host "Starting frontend -> http://localhost:5173"
Start-Process -FilePath "cmd.exe" `
    -ArgumentList "/c npm run dev" `
    -WorkingDirectory $frontend `
    -WindowStyle Hidden

Start-Sleep -Seconds 4

# ── VERIFY ────────────────────────────────────────────────────────
try {
    $h = Invoke-RestMethod -Uri "http://localhost:8081/health" -TimeoutSec 5
    Write-Host "Backend  OK — status: $($h.status)"
} catch {
    Write-Warning "Backend did not respond"
}

try {
    $null = Invoke-WebRequest -Uri "http://localhost:5173" -UseBasicParsing -TimeoutSec 5
    Write-Host "Frontend OK — http://localhost:5173"
} catch {
    Write-Warning "Frontend did not respond"
}

Write-Host ""
Write-Host "To stop everything: .\run.ps1 stop"
