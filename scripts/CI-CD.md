# CI/CD Integration Examples

This document shows how to integrate these scripts into your CI/CD pipeline.

## GitHub Actions

Create `.github/workflows/ci.yml`:

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  build-and-test:
    runs-on: windows-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build application
        shell: powershell
        run: ./scripts/build.ps1
      
      - name: Run tests with coverage
        shell: powershell
        run: ./scripts/test.ps1 -c
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests
          name: codecov-umbrella
      
      - name: Security scanning
        shell: powershell
        run: ./scripts/security-test.ps1
        continue-on-error: true

  # Optional: Release on tag
  release:
    needs: build-and-test
    runs-on: windows-latest
    if: startsWith(github.ref, 'refs/tags/')
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build release
        shell: powershell
        run: ./scripts/build.ps1
      
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: healthcare-api
          path: healthcare-api.exe
```

## Azure Pipelines

Create `azure-pipelines.yml`:

```yaml
trigger:
  - main
  - develop

pr:
  - main
  - develop

pool:
  vmImage: 'windows-latest'

variables:
  GO_VERSION: '1.21'

stages:
  - stage: Build
    displayName: 'Build and Test'
    jobs:
      - job: BuildAndTest
        displayName: 'Build, Test, and Scan'
        steps:
          - task: GoTool@0
            inputs:
              version: $(GO_VERSION)
          
          - task: PowerShell@2
            displayName: 'Build Application'
            inputs:
              filePath: $(System.DefaultWorkingDirectory)/scripts/build.ps1
          
          - task: PowerShell@2
            displayName: 'Run Tests with Coverage'
            inputs:
              filePath: $(System.DefaultWorkingDirectory)/scripts/test.ps1
              arguments: '-c'
          
          - task: PublishCodeCoverageResults@1
            displayName: 'Publish Coverage'
            inputs:
              codeCoverageTool: Cobertura
              summaryFileLocation: coverage.out
          
          - task: PowerShell@2
            displayName: 'Security Scan'
            inputs:
              filePath: $(System.DefaultWorkingDirectory)/scripts/security-test.ps1
            continueOnError: true
          
          - task: PublishBuildArtifacts@1
            displayName: 'Publish Artifact'
            inputs:
              pathToPublish: healthcare-api.exe
              artifactName: healthcare-api-$(Build.BuildId)

  - stage: Release
    displayName: 'Release'
    condition: and(succeeded(), startsWith(variables['Build.SourceBranch'], 'refs/tags/'))
    dependsOn: Build
    jobs:
      - job: CreateRelease
        displayName: 'Create Release'
        steps:
          - task: DownloadBuildArtifacts@0
            inputs:
              buildType: current
          
          - task: GitHubRelease@1
            inputs:
              gitHubConnection: github-connection
              repositoryName: your-org/healthcare-api
              action: create
              target: $(Build.SourceVersion)
              tagSource: gitTag
              assetUploadMode: replace
              assets: healthcare-api.exe
```

## GitLab CI

Create `.gitlab-ci.yml`:

```yaml
stages:
  - build
  - test
  - security
  - release

variables:
  GO_VERSION: "1.21"

build:
  stage: build
  image: golang:1.21-windowsservercore
  script:
    - powershell -Command "& './scripts/build.ps1'"
  artifacts:
    paths:
      - healthcare-api.exe
    expire_in: 1 day

test:
  stage: test
  image: golang:1.21-windowsservercore
  script:
    - powershell -Command "& './scripts/test.ps1' '-c'"
  coverage: '/coverage: \d+\.\d+%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.out

security:
  stage: security
  image: golang:1.21-windowsservercore
  script:
    - powershell -Command "& './scripts/security-test.ps1'"
  allow_failure: true

release:
  stage: release
  image: golang:1.21-windowsservercore
  script:
    - powershell -Command "& './scripts/build.ps1'"
  only:
    - tags
  artifacts:
    paths:
      - healthcare-api.exe
    expire_in: 30 days
```

## Docker Build

Create `Dockerfile.build`:

```dockerfile
FROM mcr.microsoft.com/windows/servercore:ltsc2022 AS builder

# Install Go
RUN powershell -Command \
    Invoke-WebRequest https://go.dev/dl/go1.21.0.windows-amd64.zip -OutFile /tmp/go.zip; \
    Expand-Archive /tmp/go.zip -DestinationPath C:/; \
    Remove-Item /tmp/go.zip

# Install git
RUN powershell -Command \
    choco install git -y

# Copy source
WORKDIR /build
COPY . .

# Build
RUN powershell -Command "& './scripts/build.ps1'"

# Result stage
FROM mcr.microsoft.com/windows/nanoserver:ltsc2022

COPY --from=builder /build/healthcare-api.exe /app/

WORKDIR /app
EXPOSE 8080

CMD ["healthcare-api.exe"]
```

## Jenkins Pipeline

Create `Jenkinsfile`:

```groovy
pipeline {
    agent {
        label 'windows-agent'
    }
    
    environment {
        GO_VERSION = '1.21'
    }
    
    stages {
        stage('Build') {
            steps {
                powershell './scripts/build.ps1'
            }
        }
        
        stage('Test') {
            steps {
                powershell './scripts/test.ps1 -c'
                publishHTML(
                    reportDir: '',
                    reportFiles: 'coverage.html',
                    reportName: 'Coverage Report'
                )
            }
        }
        
        stage('Security') {
            steps {
                powershell './scripts/security-test.ps1'
            }
            post {
                always {
                    step([$class: 'JunitResultArchiver',
                          testResults: 'test-results.xml'])
                }
            }
        }
        
        stage('Deploy') {
            when {
                branch 'main'
            }
            steps {
                powershell '''
                    # Deploy logic here
                    Copy-Item -Path healthcare-api.exe -Destination '\\\\deployment\\\\server\\'
                '''
            }
        }
    }
    
    post {
        always {
            cleanWs()
        }
        failure {
            emailext(
                to: '${DEFAULT_RECIPIENTS}',
                subject: 'Build failed: $PROJECT_NAME',
                body: 'See ${BUILD_URL} for details'
            )
        }
    }
}
```

## Local CI Simulation

Test your CI locally before pushing:

```powershell
# Simulate CI pipeline locally
Write-Host "=== Simulating CI Pipeline ===" -ForegroundColor Cyan

Write-Host "`nStep 1: Build" -ForegroundColor Yellow
.\scripts\build.ps1
if ($LASTEXITCODE -ne 0) { exit 1 }

Write-Host "`nStep 2: Test with Coverage" -ForegroundColor Yellow
.\scripts\test.ps1 -c
if ($LASTEXITCODE -ne 0) { exit 1 }

Write-Host "`nStep 3: Security Scan" -ForegroundColor Yellow
.\scripts\security-test.ps1 -WarningAction SilentlyContinue

Write-Host "`n=== All stages passed! ===" -ForegroundColor Green

# Generate report
if (Test-Path coverage.html) {
    Write-Host "Coverage report ready: coverage.html"
}
```

## Coverage Integration

### Codecov Setup

```yaml
# codecov.yml
coverage:
  precision: 2
  round: down
  range: "70...100"

ignore:
  - "tests"
  - "mock"
  - "vendor"
```

### Badge in README

```markdown
[![codecov](https://codecov.io/gh/your-org/healthcare-api/branch/main/graph/badge.svg?token=YOUR_TOKEN)](https://codecov.io/gh/your-org/healthcare-api)
```

## Troubleshooting CI

### Tests fail in CI but pass locally

1. Clear cache: `go clean -testcache`
2. Match Go version: Check CI uses same Go version as local
3. Add verbose output: `./scripts/test.ps1 -v`

### Artifacts not uploading

1. Check file paths are correct
2. Verify permissions on build agent
3. Check artifact retention policy

### PowerShell execution policy error in CI

Add to pipeline:
```yaml
- name: Set PowerShell policy
  shell: powershell
  run: Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process
```

---

## Security Best Practices

1. **Never commit secrets** - use CI/CD variables
2. **Sign releases** - use GPG keys
3. **Run security scans** on every commit
4. **Update dependencies** regularly
5. **Lock Go version** in CI/CD

## Performance Tips

1. **Cache Go modules**: CI systems cache automatically
2. **Parallel jobs**: Run multiple jobs in parallel
3. **Skip tests on docs**: Add path filters
4. **Use lightweight containers**: Alpine for non-Windows builds

---

For detailed script documentation, see [README.md](README.md)
