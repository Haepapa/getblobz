# Quick Start

## Installation

```bash
curl -sL https://raw.githubusercontent.com/haepapa/getblobz/main/install.sh | bash
```

Or download a binary from the [releases page](https://github.com/haepapa/getblobz/releases).

## First Sync

```bash
getblobz sync \
  --container mycontainer \
  --connection-string "DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;" \
  --output-path ./downloads \
  --disk-warn-percent 80 --disk-stop-percent 90
```

## Using a Config File

```bash
# Create config template
getblobz init

# Edit getblobz.yaml
azure:
  connection_string: "..."
sync:
  container: "mycontainer"
  output_path: "./downloads"

# Run sync
getblobz sync
```

## Common Examples

**Watch mode** (continuous sync):
```bash
getblobz sync --container mycontainer --connection-string "..." --watch --watch-interval 5m
```

**Sync specific prefix**:
```bash
getblobz sync --container mycontainer --connection-string "..." --prefix "data/2024/"
```

**Organize large file collections**:
```bash
getblobz sync --container mycontainer --connection-string "..." \
  --organize-folders --max-files-per-folder 10000
```

**Low-power device** (Raspberry Pi):
```bash
getblobz sync --container mycontainer --connection-string "..." \
  --workers 2 --max-cpu-percent 50
```

## Check Status

```bash
getblobz status
```

## Authentication

**Connection string** (most common):
```bash
--connection-string "DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;"
```

**Managed identity** (on Azure VM):
```bash
--account-name myaccount --use-managed-identity
```

**Azure CLI**:
```bash
az login
--account-name myaccount --use-azure-cli
```

## Environment Variables

```bash
export GETBLOBZ_CONNECTION_STRING="..."
export GETBLOBZ_CONTAINER=mycontainer
export GETBLOBZ_OUTPUT_PATH=./downloads
getblobz sync
```

## Get Help

```bash
getblobz --help
getblobz sync --help
```

See [README.md](README.md) for more details.
