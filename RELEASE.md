# Release and Deployment Guide

This document describes the release process, versioning strategy, and deployment options for getblobz.

## Table of Contents

- [Overview](#overview)
- [Versioning Strategy](#versioning-strategy)
- [Branch Strategy](#branch-strategy)
- [Release Process](#release-process)
- [Distribution Methods](#distribution-methods)
- [For Developers](#for-developers)
- [Troubleshooting](#troubleshooting)

---

## Overview

getblobz uses an automated CI/CD pipeline to build, test, and release software. Releases are triggered by merging code into specific branches:

- **`test` branch**: Pre-release/staging environment
- **`prod` branch**: Production releases
- **`v*` tags**: Versioned releases

## Versioning Strategy

We follow [Semantic Versioning 2.0.0](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backwards compatible)
- **PATCH**: Bug fixes (backwards compatible)

### Version Examples

- `v1.0.0` - Initial stable release
- `v1.1.0` - New features added
- `v1.1.1` - Bug fixes
- `v2.0.0` - Breaking changes

### Pre-release Versions

Test releases include additional identifiers:

- `v1.0.0-test.abc1234` - Test release from commit `abc1234`
- `v1.0.0-rc.1` - Release candidate 1

---

## Branch Strategy

### Main Branches

```
main (development)
  ↓
test (staging/pre-production)
  ↓
prod (production)
```

### Workflow

1. **Development**: Work on feature branches, merge to `main`
2. **Testing**: Merge `main` → `test` for integration testing
3. **Production**: Merge `test` → `prod` for production release

### Branch Protection

**`prod` branch** should have:
- Required reviews (2+ approvals)
- Required status checks
- No force pushes
- No deletions

**`test` branch** should have:
- Required reviews (1+ approval)
- Required status checks

---

## Release Process

### Automated Release (Recommended)

#### For Test Releases

```bash
# Ensure you're on main with latest changes
git checkout main
git pull origin main

# Merge to test branch
git checkout test
git pull origin test
git merge main

# Push to trigger release
git push origin test
```

This automatically:
1. Runs all tests
2. Builds binaries for all platforms
3. Creates Docker images
4. Creates a pre-release on GitHub
5. Tags version as `vX.Y.Z-test.COMMIT`

#### For Production Releases

```bash
# Ensure test is stable
git checkout test
git pull origin test

# Merge to prod
git checkout prod
git pull origin prod
git merge test

# Push to trigger release
git push origin prod
```

This automatically:
1. Runs all tests
2. Builds binaries for all platforms
3. Creates Docker images (tagged as `latest`)
4. Creates a full GitHub release
5. Uses last git tag or increments version

### Manual Tagged Release

For specific version releases:

```bash
# Create and push tag
git tag -a v1.2.3 -m "Release version 1.2.3"
git push origin v1.2.3
```

This triggers the same release pipeline with the specific version.

---

## Distribution Methods

### 1. Pre-built Binaries

**Recommended for most users**

Download from [GitHub Releases](https://github.com/haepapa/getblobz/releases):

#### Linux (amd64)
```bash
# Download
curl -L -o getblobz https://github.com/haepapa/getblobz/releases/download/v1.0.0/getblobz-v1.0.0-linux-amd64

# Make executable
chmod +x getblobz

# Move to PATH
sudo mv getblobz /usr/local/bin/

# Verify
getblobz --version
```

#### Linux (arm64 - Raspberry Pi, AWS Graviton)
```bash
curl -L -o getblobz https://github.com/haepapa/getblobz/releases/download/v1.0.0/getblobz-v1.0.0-linux-arm64
chmod +x getblobz
sudo mv getblobz /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L -o getblobz https://github.com/haepapa/getblobz/releases/download/v1.0.0/getblobz-v1.0.0-darwin-amd64
chmod +x getblobz
sudo mv getblobz /usr/local/bin/
```

#### macOS (Apple Silicon - M1/M2)
```bash
curl -L -o getblobz https://github.com/haepapa/getblobz/releases/download/v1.0.0/getblobz-v1.0.0-darwin-arm64
chmod +x getblobz
sudo mv getblobz /usr/local/bin/
```

#### Windows (PowerShell)
```powershell
# Download
Invoke-WebRequest `
  -Uri "https://github.com/haepapa/getblobz/releases/download/v1.0.0/getblobz-v1.0.0-windows-amd64.exe" `
  -OutFile "getblobz.exe"

# Move to PATH (adjust path as needed)
Move-Item getblobz.exe C:\Windows\System32\
```

#### Verification

Verify checksums:
```bash
# Download checksum file
curl -L -O https://github.com/haepapa/getblobz/releases/download/v1.0.0/getblobz-v1.0.0-linux-amd64.sha256

# Verify
sha256sum -c getblobz-v1.0.0-linux-amd64.sha256
```

### 2. Docker Container

**Recommended for containerized environments**

#### Using Docker

```bash
# Pull image
docker pull ghcr.io/haepapa/getblobz:latest

# Run with volume mounts
docker run --rm \
  -v $(pwd)/data:/data \
  -e GETBLOBZ_CONNECTION_STRING="your-connection-string" \
  ghcr.io/haepapa/getblobz:latest sync \
  --container mycontainer

# With config file
docker run --rm \
  -v $(pwd)/config:/config \
  -v $(pwd)/data:/data \
  ghcr.io/haepapa/getblobz:latest sync \
  --config /config/getblobz.yaml
```

#### Using Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  getblobz:
    image: ghcr.io/haepapa/getblobz:latest
    volumes:
      - ./data:/data
      - ./config:/config
    environment:
      - GETBLOBZ_CONNECTION_STRING=${AZURE_CONNECTION_STRING}
      - GETBLOBZ_CONTAINER=mycontainer
    command: sync --watch --watch-interval 5m
```

Run:
```bash
docker-compose up -d
```

#### Available Tags

- `latest` - Latest production release
- `test` - Latest test release
- `v1.0.0` - Specific version
- `v1.0` - Latest patch for minor version
- `v1` - Latest minor for major version

### 3. Build from Source

**For developers or custom builds**

#### Prerequisites

- Go 1.21 or higher
- Git
- C compiler (for SQLite)
  - Linux: `gcc`
  - macOS: Xcode Command Line Tools
  - Windows: MinGW-w64

#### Clone and Build

```bash
# Clone repository
git clone https://github.com/haepapa/getblobz.git
cd getblobz

# Checkout specific version (optional)
git checkout v1.0.0

# Download dependencies
go mod download

# Build
go build -o getblobz main.go

# Install
sudo mv getblobz /usr/local/bin/

# Verify
getblobz --version
```

#### Build with Version Info

```bash
VERSION="v1.0.0"
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

go build \
  -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
  -o getblobz \
  main.go
```

#### Cross-Compilation

```bash
# For Linux
GOOS=linux GOARCH=amd64 go build -o getblobz-linux main.go

# For Windows
GOOS=windows GOARCH=amd64 go build -o getblobz-windows.exe main.go

# For macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o getblobz-darwin-arm64 main.go
```

---

## For Developers

### Pre-Release Checklist

Before merging to `test` or `prod`:

- [ ] All tests pass locally (`make test-all`)
- [ ] Code is formatted (`make fmt`)
- [ ] Linting passes (`make lint`)
- [ ] Documentation updated (README, CHANGELOG)
- [ ] Version bumped (if manual tagging)
- [ ] Breaking changes documented

### Creating a New Release

#### 1. Update Version

If using manual versioning, update in relevant files:
- Git tag (primary source of truth)
- CHANGELOG.md

#### 2. Update CHANGELOG

```markdown
## [1.2.0] - 2024-11-22

### Added
- New feature X
- Support for Y

### Changed
- Improved performance of Z

### Fixed
- Bug in component A

### Deprecated
- Feature B (will be removed in v2.0.0)
```

#### 3. Merge to Test

```bash
git checkout main
git pull origin main

# Create release branch (optional)
git checkout -b release/v1.2.0

# Update CHANGELOG.md
# Commit changes

git checkout test
git merge release/v1.2.0
git push origin test
```

#### 4. Verify Test Release

- Check GitHub Actions for successful build
- Download and test binaries
- Pull and test Docker image
- Verify release notes

#### 5. Promote to Production

```bash
git checkout prod
git merge test
git push origin prod

# Optional: Create tag
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0
```

### GitHub Actions Workflow

The release process is automated in `.github/workflows/release.yml`:

```
┌─────────────┐
│   Trigger   │  Push to test/prod or tag v*
└──────┬──────┘
       │
       ├─────────────────────────────────┐
       │                                 │
┌──────▼──────┐                   ┌─────▼────────┐
│ Prepare     │                   │ Run Tests    │
│ Version     │                   │              │
└──────┬──────┘                   └─────┬────────┘
       │                                 │
       └─────────────┬───────────────────┘
                     │
       ┌─────────────┴─────────────┐
       │                           │
┌──────▼────────┐          ┌───────▼────────┐
│ Build         │          │ Build Docker   │
│ Binaries      │          │ Image          │
│ (Multi-arch)  │          │ (Multi-arch)   │
└──────┬────────┘          └───────┬────────┘
       │                           │
       └─────────────┬─────────────┘
                     │
              ┌──────▼───────┐
              │ Create       │
              │ GitHub       │
              │ Release      │
              └──────────────┘
```

### Rollback Procedure

If a release has issues:

#### Rollback Binary Release

1. Mark release as draft in GitHub
2. Users downloading `latest` will get previous version
3. Create hotfix and new release

#### Rollback Docker Image

```bash
# Re-tag previous version as latest
docker pull ghcr.io/haepapa/getblobz:v1.1.0
docker tag ghcr.io/haepapa/getblobz:v1.1.0 ghcr.io/haepapa/getblobz:latest
docker push ghcr.io/haepapa/getblobz:latest
```

#### Revert Git Changes

```bash
# On prod branch
git checkout prod
git revert HEAD
git push origin prod
```

### Release Artifacts

Each release generates:

**Binaries** (per platform):
- Binary file (`getblobz-VERSION-OS-ARCH`)
- SHA256 checksum (`.sha256` file)

**Docker Images**:
- `ghcr.io/haepapa/getblobz:VERSION`
- `ghcr.io/haepapa/getblobz:latest` (prod only)
- `ghcr.io/haepapa/getblobz:test` (test only)

**Documentation**:
- Release notes (auto-generated)
- CHANGELOG excerpt
- Installation instructions

### Environment Variables

CI/CD workflow uses these secrets (configure in GitHub):

- `GITHUB_TOKEN` - Automatically provided by GitHub Actions
- Additional secrets (if needed):
  - `DOCKERHUB_USERNAME` - For Docker Hub publishing
  - `DOCKERHUB_TOKEN` - For Docker Hub publishing

---

## Troubleshooting

### Build Fails on GitHub Actions

**Check logs**:
1. Go to GitHub Actions tab
2. Click on failed workflow
3. Review step logs

**Common issues**:
- Test failures: Fix tests before merging
- Dependency issues: Update `go.mod`
- Platform-specific issues: Test locally with cross-compilation

### Docker Image Won't Build

```bash
# Build locally to test
docker build -t getblobz:test .

# Check for issues
docker run --rm getblobz:test --version
```

### Version Not Updating

Ensure version info is passed via ldflags:

```bash
go build -ldflags="-X main.version=v1.0.0" -o getblobz main.go
./getblobz --version
```

### Release Not Created

Check that:
- GitHub Actions has write permissions
- Branch is `test` or `prod`
- Tests passed
- Workflow completed

### Binary Doesn't Run

**Linux**: Check if executable:
```bash
chmod +x getblobz
```

**macOS**: Remove quarantine:
```bash
xattr -d com.apple.quarantine getblobz
```

**Windows**: Check antivirus/SmartScreen

---

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Documentation](https://docs.docker.com/)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)

---

## Support

For issues with releases:
- GitHub Issues: https://github.com/haepapa/getblobz/issues
- Discussions: https://github.com/haepapa/getblobz/discussions

---

**Document Version**: 1.0  
**Last Updated**: 2024-11-22  
**Maintained By**: Release Engineering Team
