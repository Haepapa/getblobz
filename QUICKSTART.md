# getblobz Quick Start Guide

## Installation

```bash
# Build from source
go build -o getblobz main.go

# Or install directly
go install
```

## First Sync

### Step 1: Create Configuration

```bash
getblobz init
```

This creates a `getblobz.yaml` file in the current directory.

### Step 2: Edit Configuration

Edit `getblobz.yaml` with your Azure Storage credentials:

```yaml
azure:
  connection_string: "DefaultEndpointsProtocol=https;AccountName=YOUR_ACCOUNT;AccountKey=YOUR_KEY;EndpointSuffix=core.windows.net"

sync:
  container: "YOUR_CONTAINER_NAME"
  output_path: "./downloads"
  workers: 10
```

### Step 3: Run Sync

```bash
getblobz sync
```

## Common Use Cases

### One-Time Sync

Download all files once:

```bash
getblobz sync \
  --container mycontainer \
  --connection-string "..." \
  --output-path ./downloads
```

### Continuous Monitoring

Watch for new files and sync automatically:

```bash
getblobz sync \
  --container mycontainer \
  --connection-string "..." \
  --watch \
  --watch-interval 5m
```

### Sync Specific Folder

Download only files with a specific prefix:

```bash
getblobz sync \
  --container mycontainer \
  --connection-string "..." \
  --prefix "data/2024/"
```

### High-Performance Sync

Use more workers for faster downloads:

```bash
getblobz sync \
  --container mycontainer \
  --connection-string "..." \
  --workers 20 \
  --batch-size 10000
```

### Resource-Constrained Sync

Limit resources on low-power devices:

```bash
getblobz sync \
  --container mycontainer \
  --connection-string "..." \
  --workers 2 \
  --max-cpu-percent 50
```

## Check Status

View sync statistics:

```bash
getblobz status
```

Output shows:
- Total sync runs
- Files downloaded/pending/failed
- Recent errors

## Authentication Methods

### Using Connection String (Recommended)

```bash
getblobz sync \
  --container mycontainer \
  --connection-string "DefaultEndpointsProtocol=https;..."
```

### Using Account Key

```bash
getblobz sync \
  --container mycontainer \
  --account-name myaccount \
  --account-key "YOUR_KEY"
```

### Using Managed Identity (Azure VM)

```bash
getblobz sync \
  --container mycontainer \
  --account-name myaccount \
  --use-managed-identity
```

### Using Azure CLI

```bash
az login
getblobz sync \
  --container mycontainer \
  --account-name myaccount \
  --use-azure-cli
```

## Environment Variables

Set credentials via environment variables:

```bash
export getblobz_CONNECTION_STRING="DefaultEndpointsProtocol=..."
export getblobz_CONTAINER=mycontainer
export getblobz_OUTPUT_PATH=./downloads

# Now just run
getblobz sync
```

## Configuration File Locations

getblobz looks for config files in this order:

1. Path specified with `--config` flag
2. `./getblobz.yaml` (current directory)
3. `~/.config/getblobz/config.yaml` (user config directory)

## Troubleshooting

### "Container not found"

- Verify container name is correct
- Check credentials have access to the container

### "Authentication failed"

- Verify connection string format
- Check account key is correct
- Ensure credentials are not expired

### Slow downloads

- Increase workers: `--workers 20`
- Check network connection
- Verify Azure region

### Database locked

- Ensure only one getblobz instance is running
- Check file permissions on `.sync-state.db`

## Next Steps

- Read [README.md](README.md) for detailed documentation
- See [DEVELOPERNOTES.md](DEVELOPERNOTES.md) for architecture details
- Check available flags: `getblobz sync --help`
