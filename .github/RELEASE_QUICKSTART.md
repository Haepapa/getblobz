# Release Quick Start

Quick reference for deploying getblobz updates.

## For End Users

### Install getblobz

**One-line install:**
```bash
curl -sL https://raw.githubusercontent.com/haepapa/getblobz/main/install.sh | bash
```

**Docker:**
```bash
docker pull ghcr.io/haepapa/getblobz:latest
```

**See [README.md](../README.md) for all installation options.**

---

## For Developers

### Deploy to Test Environment

```bash
git checkout test
git pull origin test
git merge main
git push origin test
```

→ Pre-release created automatically at https://github.com/haepapa/getblobz/releases

### Deploy to Production

```bash
git checkout prod
git pull origin prod
git merge test
git push origin prod
```

→ Full release created automatically

### Emergency Hotfix

```bash
git checkout prod
git checkout -b hotfix/issue-name
# ... make fix ...
git checkout prod && git merge hotfix/issue-name && git push
git checkout test && git merge hotfix/issue-name && git push
git checkout main && git merge hotfix/issue-name && git push
```

### Local Testing

```bash
make test-all          # Run all tests
make build-release     # Build with version
make docker-build      # Build Docker image
./getblobz --version   # Verify version
```

---

## Branch Strategy

- `main` - Development (no auto-release)
- `test` - Staging (creates pre-release)
- `prod` - Production (creates full release)

---

## Release Artifacts

Each release includes:
- Binaries for 6 platforms (Linux, macOS, Windows)
- Docker images (multi-arch)
- SHA256 checksums
- Auto-generated release notes

---

## Documentation

- **DEPLOYMENT.md** - Complete deployment workflow
- **RELEASE.md** - Detailed release process
- **CHANGELOG.md** - Version history

---

## Support

- Issues: https://github.com/haepapa/getblobz/issues
- Releases: https://github.com/haepapa/getblobz/releases
- Actions: https://github.com/haepapa/getblobz/actions
