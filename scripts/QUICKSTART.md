# Healthcare API - Quick Reference

## 🚀 One-Line Setup

```powershell
cd c:\Users\askar\My_Project\invest\scripts
.\build.ps1
```

## ⚡ Essential Commands

| Command | Purpose |
|---------|---------|
| `.\build.ps1` | Build application → `healthcare-api.exe` |
| `.\test.ps1` | Run all tests |
| `.\test.ps1 -c` | Tests + coverage report |
| `.\run.ps1` | Build & start server on `:8080` |
| `.\security-test.ps1` | Full security scan |

## 📋 Script Flags

### test.ps1
```powershell
.\test.ps1                 # Basic tests
.\test.ps1 -c              # With coverage
.\test.ps1 -v              # Verbose
.\test.ps1 -s              # Short suite
.\test.ps1 -c -v           # Coverage + verbose
```

## 🔧 Typical Workflow

```powershell
# 1. Build
.\build.ps1

# 2. Test
.\test.ps1 -c

# 3. Verify security
.\security-test.ps1

# 4. Run server
.\run.ps1
```

## ✅ Verification Checklist

```powershell
# Check Go
go version

# Check Git
git --version

# Check PowerShell
$PSVersionTable.PSVersion

# Quick test
.\test.ps1 -s
```

## 📊 Output Locations

| File | Created by | Purpose |
|------|-----------|---------|
| `healthcare-api.exe` | build.ps1 | Executable binary |
| `coverage.html` | test.ps1 -c | Coverage report |
| `.out` files | test.ps1 | Temp coverage data |

## 🐛 Quick Fixes

| Problem | Solution |
|---------|----------|
| Scripts won't run | `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser` |
| Go not found | `go version` - install from golang.org |
| Port in use | Kill: `Get-Process healthcare-api \| Stop-Process -Force` |
| Tests fail | Try: `go mod tidy && go mod download` |

## 📚 Full Documentation

See [README.md](README.md) for:
- Detailed script usage
- Prerequisites setup
- Troubleshooting guide
- Advanced configuration
- CI/CD integration examples

---

**Location**: `c:\Users\askar\My_Project\invest\scripts\`  
**Go Version Required**: 1.19+  
**PowerShell**: 5.0+ recommended
