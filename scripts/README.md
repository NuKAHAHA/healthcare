# Healthcare API - Windows Scripts Documentation

> **Note**: These scripts are designed for Windows PowerShell 5.0+ and Windows Terminal.

## Overview

This folder (`scripts/`) contains Windows PowerShell scripts to build, test, and run the Healthcare API application with ease.

| Script | Purpose |
|--------|---------|
| `build.ps1` | Builds the application with version info and optimizations |
| `test.ps1` | Runs comprehensive tests with optional coverage reporting |
| `security-test.ps1` | Runs security checks and vulnerability scanning |
| `run.ps1` | Compiles and runs the server |

## Prerequisites

Before using these scripts, ensure you have:

1. **Go 1.19+** installed ([download](https://golang.org/dl))
   - Verify: `go version`

2. **Git** installed (for version tagging)
   - Verify: `git --version`

3. **PowerShell 5.0+** (comes with Windows 10+)
   - Verify: `$PSVersionTable.PSVersion`

4. **Windows Terminal** (recommended) - [download](https://github.com/microsoft/terminal)

## Quick Start

### 1. Build the Application

```powershell
cd scripts
.\build.ps1
```

**What it does:**
- Formats your code with `go fmt`
- Downloads and verifies dependencies
- Runs all tests
- Compiles the binary with version info
- Creates `healthcare-api.exe`

**Output:**
```
[SUCCESS] Build completed!
```

### 2. Run the Server

```powershell
.\run.ps1
```

**What it does:**
- Builds the application (if needed)
- Sets environment variables
- Starts the server on `http://localhost:8080`

**Output:**
```
Server starting on http://localhost:8080
```

### 3. Run Tests

```powershell
# Run all tests
.\test.ps1

# With coverage report
.\test.ps1 -c

# Verbose output
.\test.ps1 -v

# Short suite only
.\test.ps1 -s
```

**Output:**
```
[PASS] All tests passed!
[OK] Coverage report: coverage.html
```

### 4. Security Testing

```powershell
.\security-test.ps1
```

**What it does:**
- Runs unit tests
- Generates code coverage report
- Checks for vulnerable dependencies (`govulncheck`)
- Runs security scanner (`gosec`)
- Validates code with `go vet`

**Output:**
```
[SUCCESS] Security testing complete!
```

## Script Details

### build.ps1

**Usage:**
```powershell
.\build.ps1
```

**Process:**
1. Verifies Go installation
2. Formats code (`go fmt`)
3. Downloads dependencies (`go mod download`)
4. Verifies dependencies (`go mod verify`)
5. Runs tests
6. Builds binary with:
   - Git version tag
   - Build timestamp
   - Embedded version information

**Output Files:**
- `healthcare-api.exe` - Executable binary

---

### test.ps1

**Usage:**
```powershell
# All variations
.\test.ps1              # Run basic tests
.\test.ps1 -c           # With coverage report
.\test.ps1 -v           # Verbose mode
.\test.ps1 -s           # Short suite
.\test.ps1 -c -v        # Coverage + verbose
```

**Flags:**
| Flag | Description |
|------|-------------|
| `-c`, `--coverage` | Generate HTML coverage report |
| `-v`, `--verbose` | Show detailed test output |
| `-s`, `--short` | Run fast tests only |
| `--html-coverage` | Generate coverage HTML only |

**Output Files:**
- `coverage.out` - Coverage data (created with `-c` flag)
- `coverage.html` - Visual coverage report (created with `-c` flag)

---

### security-test.ps1

**Usage:**
```powershell
.\security-test.ps1
```

**Checks:**
1. ✅ Unit tests pass
2. ✅ Code coverage
3. ✅ Vulnerable dependencies (`govulncheck`)
4. ✅ Security issues (`gosec`)
5. ✅ Code quality (`go vet`)

**Tools Used:**
- `govulncheck` - Vulnerability database check
- `gosec` - Security scanner by SecureCodeWarrior

**Automatic Installation:**
Scripts will auto-install missing tools:
```powershell
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
```

**Output Files:**
- `coverage.html` - Test coverage report

---

### run.ps1

**Usage:**
```powershell
.\run.ps1
```

**Process:**
1. Builds the application
2. Sets environment variables:
   - `ENV_NAME=development`
   - `SERVER_PORT=8080`
3. Starts the server
4. Runs until Ctrl+C

**Output:**
```
Starting Healthcare API server...
Build successful!
Server starting on http://localhost:8080
```

---

## Troubleshooting

### PowerShell Execution Policy Error

If you get: `cannot be loaded because running scripts is disabled`

**Solution:**
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

Then confirm with `Y`.

### "Go is not installed"

**Solution:**
1. Download Go: https://golang.org/dl
2. Install and restart PowerShell
3. Verify: `go version`

### Build Fails: "Cannot find package"

**Solution:**
```powershell
go mod tidy
go mod download
.\build.ps1
```

### Tests Fail: "Port already in use"

**Solution:**
```powershell
# Kill existing process
Get-Process healthcare-api -ErrorAction SilentlyContinue | Stop-Process -Force

# Or change port in run.ps1
```

### Coverage Report Won't Open

**Solution:**
```powershell
# Open in default browser
Start-Process coverage.html

# Or in specific browser
Start-Process -FilePath "C:\Program Files\...\chrome.exe" -ArgumentList "coverage.html"
```

## Advanced Usage

### Build with Custom Version

Edit `build.ps1` and modify:
```powershell
$version = "2.1.0-beta"
```

### Run with Custom Port

Edit `run.ps1` and change:
```powershell
$env:SERVER_PORT = "9000"
```

### CI/CD Integration

Use in GitHub Actions or Azure Pipelines:
```yaml
- name: Build
  run: .\scripts\build.ps1

- name: Test
  run: .\scripts\test.ps1 -c

- name: Security Check
  run: .\scripts\security-test.ps1
```

## File Structure

```
scripts/
├── build.ps1              # Build the application
├── test.ps1               # Run tests with coverage
├── security-test.ps1      # Security scanning
├── run.ps1                # Run the server
└── README.md              # This file
```

## Common Commands Cheat Sheet

```powershell
# Development workflow
.\build.ps1            # Build everything
.\test.ps1 -c          # Test with coverage
.\run.ps1              # Start server

# Continuous testing (watch mode)
.\test.ps1 -v          # Watch and rebuild

# Pre-deployment
.\security-test.ps1    # Full security check
.\build.ps1            # Final build

# Cleanup (if needed)
rm healthcare-api.exe, coverage.out, coverage.html
```

## Performance Tips

1. **First run is slow** (downloads deps) - subsequent runs are fast
2. **Cache Go modules**: Scripts cache dependencies automatically
3. **Use short tests** for quick feedback: `.\test.ps1 -s`
4. **Run tests in parallel**: Go does this automatically

## Support

For issues:
1. Check the error message carefully
2. Run with verbose flags: `.\test.ps1 -v`
3. Verify prerequisites: `go version`, `git --version`
4. Check system logs in `%AppData%\Code\User\workspaceStorage\`

## License

These scripts are provided as-is for the Healthcare API project.

---

**Last Updated**: 2024
**Compatibility**: Windows 10+, PowerShell 5.0+, Go 1.19+
