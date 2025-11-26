# Release Process

## Versioning

We use [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

- **MAJOR**: Breaking changes
- **MINOR**: New features (backwards compatible)
- **PATCH**: Bug fixes

## Creating a Release

### 1. Update Version

Update version in relevant files and commit:

```bash
git commit -m "Bump version to v1.2.0"
git push origin main
```

### 2. Merge to Test

```bash
git checkout test
git merge main
git push origin test
```

Wait for tests to pass and validate in test environment.

### 3. Merge to Prod

```bash
git checkout prod
git merge test
git push origin prod
```

### 4. Tag the Release

```bash
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0
```

This triggers the release workflow which:
- Builds binaries for all platforms
- Creates Docker images
- Publishes GitHub release

## Pre-release Checklist

- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped
- [ ] Tested in staging environment

## Distribution

Releases are automatically published to:
- GitHub Releases (binaries)
- GitHub Container Registry (Docker images)

## Install Script

Update version in `install.sh` if needed.

## Rollback

If a release has issues:

```bash
git revert <commit-sha>
git push origin prod
```

Or tag a previous version as latest.
