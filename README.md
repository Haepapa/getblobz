# getblobz

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A CLI tool for syncing files from Azure Blob Storage to your local filesystem. Handles large numbers of files efficiently and works on everything from Raspberry Pi to high-performance servers.

## Features

- Fast incremental sync with state tracking
- Multiple Azure auth methods (connection string, managed identity, Azure CLI)
- Robust resume, checksums, and optional watch mode

## Installation

### Quick Install

```bash
curl -sL https://raw.githubusercontent.com/haepapa/getblobz/main/install.sh | bash
```

### Download Binary

Download from the [releases page](https://github.com/haepapa/getblobz/releases):

```bash
# Linux
curl -L -o getblobz https://github.com/haepapa/getblobz/releases/latest/download/getblobz-linux-amd64
chmod +x getblobz
sudo mv getblobz /usr/local/bin/

# macOS
curl -L -o getblobz https://github.com/haepapa/getblobz/releases/latest/download/getblobz-darwin-arm64
chmod +x getblobz
sudo mv getblobz /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/haepapa/getblobz.git
cd getblobz
go build -o getblobz main.go
```

See [docs/BUILD.md](docs/BUILD.md) for more options.

## Quick Start

```bash
# Basic sync
getblobz sync \
  --container mycontainer \
  --connection-string "DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;" \
  --output-path ./downloads

# Watch mode
getblobz sync --container mycontainer --connection-string "..." --watch --watch-interval 5m
```

More examples are in docs/README.md.

## Commands

- `sync` - Sync blobs from Azure Storage to local filesystem
- `init` - Generate configuration file template
- `status` - Show sync statistics

Run `getblobz <command> --help` for detailed options.

## Authentication

```bash
# Connection string
--connection-string "DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;"

# Managed identity (Azure VM/Container)
--account-name myaccount --use-managed-identity

# Azure CLI
az login
--account-name myaccount --use-azure-cli
```

## Configuration

Create a config file:

```bash
getblobz init
```

Edit `getblobz.yaml`:

```yaml
azure:
  connection_string: "DefaultEndpointsProtocol=https;..."

sync:
  container: "mycontainer"
  output_path: "./downloads"
  workers: 10
  disk_warn_percent: 80
  disk_stop_percent: 90

watch:
  enabled: false
  interval: "5m"
```

Or use environment variables with `GETBLOBZ_` prefix:

```bash
export GETBLOBZ_CONTAINER=mycontainer
export GETBLOBZ_CONNECTION_STRING="..."
```

## Folder Organization

For large file collections, enable folder organization to maintain filesystem performance:

```bash
# Sequential strategy (fills folders in order)
--organize-folders --max-files-per-folder 10000 --folder-strategy sequential

# Partition key (hash-based distribution, good for Spark)
--organize-folders --folder-strategy partition_key --partition-depth 2

# Date-based (organizes by download date)
--organize-folders --folder-strategy date
```





## Documentation

- Detailed guides live in the docs/ folder: Build, Testing, Deployment, Release, and Developer.
- Quick Start is covered above; see docs/TESTING.md for Azurite setup if running integration/E2E tests.
## Contributing

Contributions welcome! Fork the repo, make changes, add tests, and submit a pull request.

## License

MIT License - see [LICENSE](LICENSE).

## Support

[GitHub Issues](https://github.com/haepapa/getblobz/issues)