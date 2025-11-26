# Deployment Guide

## Branch Strategy

```
main (development)
  ↓
test (staging)
  ↓
prod (production)
```

## Standard Deployment

```bash
# 1. Merge to test branch
git checkout test
git merge main
git push origin test
# → Triggers test deployment

# 2. Test in staging environment

# 3. Deploy to production
git checkout prod
git merge test
git push origin prod
# → Triggers production release
```

## Hotfix Process

```bash
# 1. Branch from prod
git checkout prod
git checkout -b hotfix/critical-bug

# 2. Fix and test
git commit -m "Fix critical bug"

# 3. Merge to prod
git checkout prod
git merge hotfix/critical-bug
git push origin prod

# 4. Backport to other branches
git checkout test && git merge hotfix/critical-bug && git push
git checkout main && git merge hotfix/critical-bug && git push
```

## Rollback

```bash
# Revert last commit
git revert HEAD
git push origin prod

# Or reset to specific commit
git reset --hard <commit-sha>
git push origin prod --force
```

## Docker

```bash
# Build
docker build -t getblobz:latest .

# Run
docker run --rm \
  -v $(pwd)/data:/data \
  -e GETBLOBZ_CONNECTION_STRING="..." \
  getblobz:latest sync --container mycontainer
```

## Systemd Service

Create `/etc/systemd/system/getblobz.service`:

```ini
[Unit]
Description=getblobz sync service
After=network.target

[Service]
Type=simple
User=getblobz
WorkingDirectory=/opt/getblobz
ExecStart=/usr/local/bin/getblobz sync --watch
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable getblobz
sudo systemctl start getblobz
```
