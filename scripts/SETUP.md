# First Time Setup Guide

## Step 1: Prerequisites Installation

### Install Go 1.19+

1. Download from: https://golang.org/dl
2. Run the installer
3. Restart PowerShell
4. Verify installation:
   ```powershell
   go version
   ```
   Expected output: `go version go1.21.x windows/amd64` (or newer)

### Install Git

1. Download from: https://git-scm.com/download/win
2. Run the installer
3. Restart PowerShell
4. Verify installation:
   ```powershell
   git --version
   ```

### Verify PowerShell

Check PowerShell version (should be 5.0+):
```powershell
$PSVersionTable.PSVersion
```

**Windows 10+** comes with PowerShell 5.0. For Windows 7/8, download from Microsoft Store.

---

## Step 2: Enable Script Execution

PowerShell may block script execution. Enable it:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

When prompted, type `Y` and press Enter.

**Verify:**
```powershell
Get-ExecutionPolicy
```
Should show: `RemoteSigned`

---

## Step 3: First Time Build

Navigate to the scripts folder:

```powershell
cd c:\Users\askar\My_Project\invest\scripts
```

Run the build script:

```powershell
.\build.ps1
```

**Expected output:**
```
================================================
Healthcare API - Build Script
================================================

[CHECK] Verifying Go installation...
OK: go version go1.21.x windows/amd64

[FORMAT] Formatting code...
OK: Code formatted

...

[SUCCESS] Build completed!

Next steps:
  1. Run: .\scripts\run.ps1
  2. Or: .\healthcare-api.exe
```

**Time for first build**: 2-5 minutes (downloading dependencies)

---

## Step 4: First Time Test

Run the test suite:

```powershell
.\test.ps1 -c
```

**Expected output:**
```
================================================
Healthcare API - Test Suite
================================================

[TEST] Running comprehensive test suite...
ok  main   0.234s

[PASS] All tests passed!
[COVERAGE] Generating coverage report...
[OK] Coverage report: coverage.html
Coverage: (coverage stats here)

[SUCCESS] Testing complete!
```

**Note**: First test run downloads test dependencies, subsequent runs are faster.

---

## Step 5: Run the Server

Start the development server:

```powershell
.\run.ps1
```

**Expected output:**
```
Starting Healthcare API server...
Build successful!

Server starting on http://localhost:8080
```

**Test the server:**

Open a new PowerShell window and test:

```powershell
Invoke-WebRequest -Uri http://localhost:8080/health -UseBasicParsing
```

You should see the health endpoint response.

**To stop the server**: Press `Ctrl+C` in the PowerShell window

---

## Step 6: View Coverage Report

After running `.\test.ps1 -c`, open the coverage report:

```powershell
Start-Process coverage.html
```

This opens the HTML coverage report in your default browser showing:
- Code coverage percentage
- Line-by-line coverage
- Uncovered code sections

---

## Quick Verification Checklist

Run this to verify everything is working:

```powershell
# Check Go
Write-Host "Go:" (go version)

# Check Git
Write-Host "Git:" (git --version)

# Check PowerShell
Write-Host "PowerShell:" $PSVersionTable.PSVersion

# Run quick tests
.\test.ps1 -s

# Show build artifact
Get-Item healthcare-api.exe | Select-Object Name, Length, LastWriteTime
```

**Expected result**: All items should show versions/info without errors

---

## Troubleshooting First Time Setup

### ❌ "Go is not installed"
- **Solution**: Download and install Go from golang.org/dl
- Restart PowerShell after installation
- Run `go version` to verify

### ❌ "PowerShell scripts cannot be run"
- **Solution**: Run this once:
  ```powershell
  Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
  ```
- Confirm with `Y`

### ❌ "Cannot find package"
- **Solution**: Ensure internet connection, then run:
  ```powershell
  go mod tidy
  go mod download
  .\build.ps1
  ```

### ❌ "Port 8080 already in use"
- **Solution**: Change the port in `run.ps1` or kill existing process:
  ```powershell
  Get-Process healthcare-api -ErrorAction SilentlyContinue | Stop-Process -Force
  ```

### ❌ "Tests timeout"
- **Solution**: This is usually network-related on first run
- Run again: `.\test.ps1 -s` (short suite)

### ❌ "Git version not found" (during build)
- **Solution**: Install Git, or edit `build.ps1` to skip version tagging

---

## Environment Variables (Optional)

The server uses these environment variables. Defaults work for development:

| Variable | Default | Purpose |
|----------|---------|---------|
| `ENV_NAME` | `development` | Environment mode |
| `SERVER_PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Logging level |
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |

To set environment variables:

```powershell
# Session only
$env:SERVER_PORT = "9000"
$env:ENV_NAME = "production"

# Permanent (machine-wide)
[Environment]::SetEnvironmentVariable("SERVER_PORT", "9000", "Machine")
```

---

## Next Steps

After first-time setup:

1. **Daily development**:
   ```powershell
   .\build.ps1          # Build
   .\test.ps1           # Quick test
   .\run.ps1            # Run server
   ```

2. **Before pushing code**:
   ```powershell
   .\security-test.ps1  # Full security scan
   ```

3. **Daily CI/CD** (if using GitHub/Azure):
   - Add scripts to your pipeline

4. **Read full documentation**:
   - See [README.md](README.md) for advanced usage
   - See [QUICKSTART.md](QUICKSTART.md) for command reference

---

## Support Resources

- **Go Documentation**: https://golang.org/doc
- **Go Modules**: https://github.com/golang/go/wiki/Modules
- **PowerShell Docs**: https://docs.microsoft.com/powershell/
- **GitHub Issues**: Check project repository

---

## Tips for Success

✅ **Do**:
- Run `.\build.ps1` after pulling new code
- Test frequently with `.\test.ps1`
- Check security before deployment
- Keep Go and dependencies updated

❌ **Don't**:
- Edit scripts without understanding them
- Skip tests before committing
- Run scripts from cmd.exe (use PowerShell)
- Keep old exe files lying around

---

**Status**: Ready to develop! 🚀

For detailed command reference, see [QUICKSTART.md](QUICKSTART.md)
