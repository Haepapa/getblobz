# Deployment Guide for Developers

This guide explains how to deploy updates to getblobz for future developers who will push changes.

## Table of Contents

- [Quick Reference](#quick-reference)
- [Branch Strategy](#branch-strategy)
- [Making Changes](#making-changes)
- [Testing Changes](#testing-changes)
- [Deploying to Test Environment](#deploying-to-test-environment)
- [Deploying to Production](#deploying-to-production)
- [Hotfixes](#hotfixes)
- [Rollback Procedures](#rollback-procedures)
- [Troubleshooting](#troubleshooting)

---

## Quick Reference

### Standard Deployment Flow

```bash
# 1. Make changes on feature branch
git checkout -b feature/my-feature
# ... make changes ...
git commit -m "Add new feature"
git push origin feature/my-feature

# 2. Create PR to main
# ... get review and merge ...

# 3. Deploy to test environment
git checkout test
git pull origin test
git merge main
git push origin test
# ⚡ Automatic deployment triggered

# 4. Verify in test environment
# ... manual testing ...

# 5. Deploy to production
git checkout prod
git pull origin prod
git merge test
git push origin prod
# ⚡ Automatic production release

```

### Emergency Hotfix

```bash
# 1. Branch from prod
git checkout prod
git pull origin prod
git checkout -b hotfix/critical-bug

# 2. Fix and commit
git commit -m "Fix critical bug"

# 3. Merge to prod
git checkout prod
git merge hotfix/critical-bug
git push origin prod

# 4. Backport to test and main
git checkout test
git merge hotfix/critical-bug
git push origin test

git checkout main
git merge hotfix/critical-bug
git push origin main
```

---

## Branch Strategy

### Branch Purpose

| Branch | Purpose | Deployment | Auto-Release |
|--------|---------|------------|--------------|
| `main` | Development | None | No |
| `test` | Staging/QA | Test Environment | Yes (pre-release) |
| `prod` | Production | Production Environment | Yes (full release) |
| `feature/*` | New features | None | No |
| `hotfix/*` | Emergency fixes | None | No |

### Branch Protection

Configure these settings in GitHub:

**`prod` branch**:
- ✅ Require pull request before merging
- ✅ Require 2 approvals
- ✅ Require status checks to pass
- ✅ Require conversation resolution
- ✅ Restrict who can push
- ✅ Do not allow force pushes
- ✅ Do not allow deletions

**`test` branch**:
- ✅ Require pull request before merging
- ✅ Require 1 approval
- ✅ Require status checks to pass
- ✅ Do not allow force pushes

---

## Making Changes

### Step 1: Create Feature Branch

```bash
# Update main branch
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/descriptive-name

# Examples:
# feature/add-retry-logic
# feature/improve-performance
# bugfix/fix-memory-leak
```

### Step 2: Develop and Test

```bash
# Make your changes
# ...

# Run local tests
make test-unit
make test-integration

# Format code
make fmt

# Run all checks
make ci

# Commit changes
git add .
git commit -m "Add descriptive commit message"
```

### Step 3: Push and Create PR

```bash
# Push feature branch
git push origin feature/descriptive-name

# Create Pull Request on GitHub
# - Title: Clear, descriptive
# - Description: What changes, why, how to test
# - Link related issues
# - Request reviewers
```

### Step 4: Address Review Comments

```bash
# Make requested changes
# ...

# Commit and push
git commit -m "Address review comments"
git push origin feature/descriptive-name

# Once approved, merge to main
# (GitHub UI or command line)
```

---

## Testing Changes

### Local Testing

Before pushing to test:

```bash
# Run full test suite
make test-all

# Build binary and test manually
make build-release
./getblobz --version
./getblobz sync --help

# Build Docker image
make docker-build
make docker-run

# Test specific scenarios
./getblobz sync --container test-container --connection-string "..."
```

### Pre-Deployment Checklist

- [ ] All tests pass locally (`make test-all`)
- [ ] Code is formatted (`make fmt`)
- [ ] Linting passes (`make lint`)
- [ ] No security vulnerabilities (`go mod tidy`)
- [ ] Documentation updated (README, CHANGELOG)
- [ ] Manual testing completed
- [ ] Breaking changes documented
- [ ] Migration guide created (if needed)

---

## Deploying to Test Environment

### Purpose

The test environment is for:
- Integration testing
- QA validation
- Stakeholder demos
- Performance testing

### Deployment Steps

#### 1. Merge to Test Branch

```bash
# Ensure main is up to date
git checkout main
git pull origin main

# Switch to test branch
git checkout test
git pull origin test

# Merge main into test
git merge main

# Resolve conflicts if any
# ...

# Push to trigger deployment
git push origin test
```

#### 2. Monitor Deployment

```bash
# Watch GitHub Actions
# https://github.com/haepapa/getblobz/actions

# Check for:
# ✅ Tests passing
# ✅ Binaries built
# ✅ Docker images created
# ✅ Pre-release created
```

#### 3. Verify Deployment

```bash
# Check release page
# https://github.com/haepapa/getblobz/releases

# Verify version
VERSION=$(git describe --tags --always --dirty)
echo "Deployed version: $VERSION"

# Download and test binary
curl -L -o getblobz-test \
  "https://github.com/haepapa/getblobz/releases/download/${VERSION}/getblobz-${VERSION}-linux-amd64"
chmod +x getblobz-test
./getblobz-test --version

# Test Docker image
docker pull ghcr.io/haepapa/getblobz:test
docker run --rm ghcr.io/haepapa/getblobz:test --version
```

#### 4. Perform QA Testing

Create a test plan and execute:

```bash
# Example test scenarios:
# 1. Fresh installation
# 2. Upgrade from previous version
# 3. Basic sync operation
# 4. Watch mode
# 5. Error handling
# 6. Configuration options
# 7. Performance under load
```

### Test Environment Variables

Set these for testing:

```bash
export GETBLOBZ_CONNECTION_STRING="..."
export GETBLOBZ_CONTAINER="test-container"
export GETBLOBZ_OUTPUT_PATH="./test-data"
```

---

## Deploying to Production

### Pre-Production Checklist

Before merging to prod:

- [ ] All changes tested in test environment
- [ ] QA sign-off received
- [ ] Stakeholder approval obtained
- [ ] CHANGELOG.md updated
- [ ] Release notes prepared
- [ ] Rollback plan ready
- [ ] Monitoring in place
- [ ] Communication plan ready

### Deployment Steps

#### 1. Final Verification

```bash
# Verify test branch is stable
git checkout test
git pull origin test

# Check latest test release
# Review GitHub Actions results
# Confirm all tests passing
```

#### 2. Merge to Production

```bash
# Switch to prod branch
git checkout prod
git pull origin prod

# Merge test into prod
git merge test

# Review merge commit
git log -1

# Push to trigger production release
git push origin prod
```

#### 3. Monitor Production Deployment

```bash
# Watch GitHub Actions
# https://github.com/haepapa/getblobz/actions

# Deployment includes:
# ✅ Full test suite
# ✅ Multi-platform binaries
# ✅ Docker images (tagged as 'latest')
# ✅ GitHub Release created
# ✅ Checksums generated
```

#### 4. Verify Production Release

```bash
# Check release is published
# https://github.com/haepapa/getblobz/releases

# Verify version
git describe --tags --abbrev=0

# Test binary download
curl -L -o getblobz \
  "https://github.com/haepapa/getblobz/releases/latest/download/getblobz-linux-amd64"
chmod +x getblobz
./getblobz --version

# Test Docker image
docker pull ghcr.io/haepapa/getblobz:latest
docker run --rm ghcr.io/haepapa/getblobz:latest --version
```

#### 5. Post-Deployment Tasks

```bash
# Update documentation
# - Deployment log
# - Known issues
# - User communications

# Monitor for issues
# - Check error logs
# - Watch for user reports
# - Monitor performance metrics

# Communicate release
# - Update announcement channels
# - Notify users of breaking changes
# - Share release notes
```

### Creating a Versioned Release

For major/minor releases, create a Git tag:

```bash
# After merging to prod
git checkout prod
git pull origin prod

# Create tag
git tag -a v1.2.0 -m "Release v1.2.0: Add feature X, fix bug Y"

# Push tag
git push origin v1.2.0

# This triggers another release with specific version
```

---

## Hotfixes

### When to Use Hotfixes

Use hotfix process for:
- Critical bugs in production
- Security vulnerabilities
- Data corruption issues
- Severe performance problems

### Hotfix Process

#### 1. Create Hotfix Branch

```bash
# Branch from prod (not main!)
git checkout prod
git pull origin prod
git checkout -b hotfix/critical-issue-description
```

#### 2. Fix the Issue

```bash
# Make minimal changes to fix the issue
# ...

# Test thoroughly
make test-all

# Commit
git commit -m "Fix critical issue: description"
```

#### 3. Deploy Hotfix

```bash
# Merge to prod immediately
git checkout prod
git merge hotfix/critical-issue-description
git push origin prod

# Deployment triggers automatically
```

#### 4. Backport to Other Branches

```bash
# Merge hotfix back to test
git checkout test
git merge hotfix/critical-issue-description
git push origin test

# Merge hotfix back to main
git checkout main
git merge hotfix/critical-issue-description
git push origin main

# Clean up hotfix branch
git branch -d hotfix/critical-issue-description
git push origin --delete hotfix/critical-issue-description
```

---

## Rollback Procedures

### Scenario 1: Bad Test Release

```bash
# Simply re-deploy from previous commit
git checkout test
git reset --hard HEAD~1  # Go back one commit
git push origin test --force
```

### Scenario 2: Bad Production Release

```bash
# Option A: Revert the merge commit
git checkout prod
git revert -m 1 HEAD
git push origin prod

# Option B: Reset to previous release (more destructive)
# Find previous good commit
git log --oneline
git reset --hard <previous-commit-sha>
git push origin prod --force  # Requires force push permissions
```

### Scenario 3: Emergency Rollback

If you need to immediately rollback:

```bash
# 1. Mark current release as draft in GitHub
#    (prevents new downloads of bad version)

# 2. Re-tag previous version
git checkout prod
git tag -f v1.1.0 HEAD~1
git push origin v1.1.0 --force

# 3. Docker: Re-tag previous image
docker pull ghcr.io/haepapa/getblobz:v1.1.0
docker tag ghcr.io/haepapa/getblobz:v1.1.0 ghcr.io/haepapa/getblobz:latest
docker push ghcr.io/haepapa/getblobz:latest
```

---

## Troubleshooting

### Deployment Fails on GitHub Actions

**Check the logs**:
1. Go to GitHub Actions tab
2. Click failed workflow
3. Review step output

**Common issues**:

```bash
# Tests failing
# → Fix tests locally first
# → Re-run workflow after fixes

# Build failing
# → Check dependency versions
# → Verify go.mod is up to date

# Docker build failing
# → Test locally: make docker-build
# → Check Dockerfile syntax
```

### Binary Won't Download

```bash
# Check release was created
curl -I https://github.com/haepapa/getblobz/releases/latest

# Check specific asset exists
curl -I https://github.com/haepapa/getblobz/releases/download/v1.0.0/getblobz-v1.0.0-linux-amd64

# Verify GitHub Actions completed successfully
```

### Docker Image Not Available

```bash
# Check if image was pushed
docker manifest inspect ghcr.io/haepapa/getblobz:latest

# Verify GITHUB_TOKEN permissions
# Settings > Actions > General > Workflow permissions
# Must have "Read and write permissions"

# Try pulling specific version
docker pull ghcr.io/haepapa/getblobz:v1.0.0
```

### Version Not Updating

```bash
# Verify version in release
./getblobz --version

# If showing "dev", check build was done with ldflags
# See Makefile: build-release target

# Rebuild with version:
VERSION=v1.0.0 make build-release
```

### Merge Conflicts

```bash
# When merging test to prod
git checkout prod
git merge test

# If conflicts occur:
git status  # See conflicted files
# Resolve conflicts manually
git add <resolved-files>
git commit
git push origin prod
```

---

## Best Practices

### Commit Messages

Use conventional commits:

```bash
feat: Add new feature X
fix: Fix bug in component Y
docs: Update README
test: Add tests for Z
chore: Update dependencies
refactor: Improve code structure
perf: Optimize performance of A
```

### Pull Request Guidelines

- **Title**: Clear and concise
- **Description**: What, why, how
- **Testing**: How to verify changes
- **Screenshots**: If UI changes
- **Breaking Changes**: Clearly marked
- **Related Issues**: Link with "Fixes #123"

### Code Review

- Review within 24 hours
- Be constructive and specific
- Test locally if possible
- Check for security issues
- Verify tests are adequate
- Ensure documentation is updated

### Communication

- Announce major releases
- Document breaking changes
- Provide migration guides
- Share release notes
- Update internal docs
- Notify stakeholders

---

## Additional Resources

- [RELEASE.md](RELEASE.md) - Detailed release documentation
- [BUILD.md](BUILD.md) - Build instructions
- [TESTERNOTES.md](TESTERNOTES.md) - Testing guidelines
- [DEVELOPERNOTES.md](DEVELOPERNOTES.md) - Developer documentation

---

**For Questions or Issues**:
- GitHub Issues: https://github.com/haepapa/getblobz/issues
- Team Channel: [your-team-channel]
- Documentation: All .md files in repository

---

**Document Version**: 1.0  
**Last Updated**: 2024-11-22  
**Maintained By**: DevOps Team
