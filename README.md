# getblobz

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A reliable, high-performance CLI tool for synchronising files from Azure Blob Storage to local filesystem. Built with Go, designed to handle large numbers of files efficiently across diverse hardware platforms.

## Features

- **Multiple Authentication Methods**: Connection string, account key, managed identity, service principal, Azure CLI
- **Incremental Sync**: Only downloads new or changed files using SQLite state tracking
- **Continuous Monitoring**: Watch mode for automatic synchronisation at intervals
- **Concurrent Downloads**: Configurable worker pool for parallel downloads
- **Resumable Operations**: Gracefully handles interruptions and resumes from last state
- **Checksum Verification**: Optional MD5 validation for data integrity
- **Flexible Configuration**: Command-line flags, configuration files, or environment variables
- **Performance Tuning**: Resource limits and auto-throttling for diverse platforms

## Installation

Choose one of three installation methods based on your needs:

### Option 1: One-Line Install Script (Recommended)

```bash
curl -sL https://raw.githubusercontent.com/haepapa/getblobz/main/install.sh | bash
```

Or with specific version:
```bash
curl -sL https://raw.githubusercontent.com/haepapa/getblobz/main/install.sh | VERSION=v1.0.0 bash
```

### Option 2: Download Pre-built Binary

Download the appropriate binary for your platform from the [releases page](https://github.com/haepapa/getblobz/releases).

**Linux (amd64)**
```bash
curl -L -o getblobz https://github.com/haepapa/getblobz/releases/download/v1.0.0/getblobz-v1.0.0-linux-amd64
chmod +x getblobz
sudo mv getblobz /usr/local/bin/
```

**macOS (Apple Silicon)**
```bash
curl -L -o getblobz https://github.com/haepapa/getblobz/releases/download/v1.0.0/getblobz-v1.0.0-darwin-arm64
chmod +x getblobz
sudo mv getblobz /usr/local/bin/
```

**Windows (PowerShell)**
```powershell
Invoke-WebRequest -Uri "https://github.com/haepapa/getblobz/releases/download/v1.0.0/getblobz-v1.0.0-windows-amd64.exe" -OutFile "getblobz.exe"
```

See [RELEASE.md](RELEASE.md) for all platform-specific instructions.

### Option 3: Docker Container

```bash
# Pull image
docker pull ghcr.io/haepapa/getblobz:latest

# Run
docker run --rm \
  -v $(pwd)/data:/data \
  -e GETBLOBZ_CONNECTION_STRING="your-connection-string" \
  ghcr.io/haepapa/getblobz:latest sync \
  --container mycontainer

# Using docker-compose
cat > docker-compose.yml << EOF
version: '3.8'
services:
  getblobz:
    image: ghcr.io/haepapa/getblobz:latest
    volumes:
      - ./data:/data
    environment:
      - GETBLOBZ_CONNECTION_STRING=\${AZURE_CONNECTION_STRING}
    command: sync --container mycontainer --watch
EOF

docker-compose up -d
```

### Option 4: Build from Source

```bash
# Prerequisites: Go 1.21+, Git, C compiler

# Clone the repository
git clone https://github.com/haepapa/getblobz.git
cd getblobz

# Build the binary
go build -o getblobz main.go

# Install to system
sudo mv getblobz /usr/local/bin/
```

See [BUILD.md](BUILD.md) for detailed build instructions.

## Quick Start

### Basic Usage

```bash
# Sync a container
getblobz sync \
  --container mycontainer \
  --connection-string "DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;" \
  --output-path ./downloads

# Watch mode for continuous sync
getblobz sync \
  --container mycontainer \
  --connection-string "..." \
  --watch \
  --watch-interval 5m
```

### Using Configuration File

Create a configuration file:

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

watch:
  enabled: true
  interval: "5m"
```

Then run:

```bash
getblobz sync
```

## Commands

### `sync`

Synchronise blobs from Azure Storage to local filesystem.

**Flags:**

```
--container string           Container name (required)
--output-path string         Local destination path (default: ./data)
--connection-string string   Azure Storage connection string
--account-name string        Storage account name
--account-key string         Storage account key
--use-managed-identity       Use Azure Managed Identity
--tenant-id string           Azure AD tenant ID
--client-id string           Azure AD client ID
--client-secret string       Azure AD client secret
--use-azure-cli              Use Azure CLI credentials
--prefix string              Only sync blobs with this prefix
--workers int                Concurrent download workers (default: 10)
--batch-size int             Blobs per listing batch (default: 5000)
--watch                      Continuous watch mode
--watch-interval duration    Check interval in watch mode (default: 5m)
--state-db string            State database path (default: ./.sync-state.db)
--force-resync               Re-download all files
--skip-existing              Skip existing files (default: true)
--verify-checksums           Verify MD5 checksums (default: true)
```

**Examples:**

```bash
# Sync with prefix filter
getblobz sync --container mycontainer --connection-string "..." --prefix "data/2024/"

# Use managed identity (Azure VM/Container)
getblobz sync --container mycontainer --account-name myaccount --use-managed-identity

# Custom worker count and batch size
getblobz sync --container mycontainer --connection-string "..." --workers 20 --batch-size 1000
```

### `init`

Generate a configuration file template.

```bash
# Generate in current directory
getblobz init

# Generate at specific path
getblobz init --config /path/to/config.yaml
```

### `status`

Display sync statistics and current state.

```bash
# Show status
getblobz status

# Status for specific database
getblobz status --state-db /path/to/.sync-state.db
```

## Authentication Methods

### Connection String

```bash
getblobz sync \
  --container mycontainer \
  --connection-string "DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;"
```

### Account Key

```bash
getblobz sync \
  --container mycontainer \
  --account-name myaccount \
  --account-key "key=="
```

### Managed Identity (Azure VM/Container)

```bash
getblobz sync \
  --container mycontainer \
  --account-name myaccount \
  --use-managed-identity
```

### Service Principal

```bash
getblobz sync \
  --container mycontainer \
  --account-name myaccount \
  --tenant-id "xxx" \
  --client-id "yyy" \
  --client-secret "zzz"
```

### Azure CLI

```bash
# First login with Azure CLI
az login

# Then use credentials
getblobz sync \
  --container mycontainer \
  --account-name myaccount \
  --use-azure-cli
```

## Configuration

### Configuration File

Configuration files are discovered in this order:

1. Explicit path via `--config` flag
2. Current directory: `./getblobz.yaml` or `./getblobz.yml`
3. User config directory: `~/.config/getblobz/config.yaml`

**Example configuration:**

```yaml
azure:
  connection_string: "DefaultEndpointsProtocol=https;..."

sync:
  container: "mycontainer"
  output_path: "./downloads"
  prefix: ""
  workers: 10
  batch_size: 5000
  skip_existing: true
  verify_checksums: true

watch:
  enabled: false
  interval: "5m"

logging:
  level: "info"
  format: "text"

state:
  database: "./.sync-state.db"

performance:
  max_memory_mb: 0
  max_cpu_percent: 80
  auto_throttle: false
  throttle_threshold: 0.8
  bandwidth_limit: ""
  disk_buffer_mb: 32
```

### Environment Variables

All configuration options can be set via environment variables with the `getblobz_` prefix:

```bash
export getblobz_CONTAINER=mycontainer
export getblobz_CONNECTION_STRING="DefaultEndpointsProtocol=..."
export getblobz_OUTPUT_PATH=./downloads
export getblobz_WORKERS=10

getblobz sync
```

### Priority Order

1. Command-line flags (highest)
2. Environment variables
3. Config file (explicit path)
4. Config file (auto-discovered)
5. Default values (lowest)

## Performance Tuning

### Resource Limits

```bash
getblobz sync \
  --container mycontainer \
  --connection-string "..." \
  --workers 4 \
  --max-memory-mb 512 \
  --max-cpu-percent 50
```

### Auto-Throttling

Enable dynamic adjustment based on system load:

```bash
getblobz sync \
  --container mycontainer \
  --connection-string "..." \
  --auto-throttle \
  --throttle-threshold 0.75
```

## State Management

getblobz uses SQLite to track sync state, enabling:

- **Incremental Sync**: Only download changed files
- **Resume Capability**: Continue from interruptions
- **Statistics**: Track downloads, failures, and history

The state database stores:
- Sync run metadata
- Individual blob states (ETag, last modified, status)
- Error logs
- Performance metrics

## Troubleshooting

### Slow Downloads

- Increase workers: `--workers 20`
- Check network bandwidth
- Verify Azure region proximity
- Disable auto-throttle if enabled

### High System Load

- Enable auto-throttle: `--auto-throttle`
- Reduce workers: `--workers 5`
- Set CPU limit: `--max-cpu-percent 50`

### Database Locked Errors

- Ensure only one instance per state database
- Check for proper cleanup on exit
- Verify file permissions

### Authentication Failures

- Verify connection string format
- Check account key is not expired
- Ensure managed identity has "Storage Blob Data Reader" role

## Development

### Building from Source

```bash
# Install dependencies
go mod download

# Build
go build -o getblobz main.go

# Run tests
go test ./...

# Build with optimisations
go build -ldflags="-s -w" -o getblobz main.go
```

### Cross-Compilation

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o getblobz-linux main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o getblobz-darwin main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o getblobz-windows.exe main.go

# ARM (Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o getblobz-arm64 main.go
```

## Architecture

```
getblobz/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── sync.go            # Sync command
│   ├── init.go            # Init command
│   └── status.go          # Status command
├── internal/              # Internal packages
│   ├── azure/            # Azure client wrapper
│   ├── config/           # Configuration management
│   ├── storage/          # State database
│   └── sync/             # Sync engine
├── pkg/                   # Public packages
│   └── logger/           # Structured logging
└── main.go               # Entry point
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/haepapa/getblobz/issues)
- **Documentation**: See [DEVELOPERNOTES.md](DEVELOPERNOTES.md) for detailed specifications

## Acknowledgments

Built with:
- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go)
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Zap](https://github.com/uber-go/zap) - Structured logging
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver