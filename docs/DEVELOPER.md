# Developer Guide

## Architecture

```
getblobz/
├── cmd/          # CLI commands (root, sync, init, status)
├── internal/     # Internal packages
│   ├── azure/    # Azure client wrapper
│   ├── config/   # Configuration management
│   ├── storage/  # SQLite state database
│   └── sync/     # Sync engine
├── pkg/          # Public packages
│   └── logger/   # Structured logging
└── main.go
```

## Core Components

### Sync Engine (`internal/sync`)

Manages the sync workflow:
1. **Discovery**: List blobs from Azure, compare with local state
2. **Download**: Worker pool downloads pending blobs concurrently
3. **Completion**: Update statistics, loop if in watch mode

### State Management (`internal/storage`)

SQLite database tracks:
- Sync run metadata
- Blob states (ETag, last modified, status)
- Error logs
- Performance metrics

### Azure Client (`internal/azure`)

Wrapper around Azure SDK providing:
- Multiple authentication methods
- Blob listing with pagination
- Concurrent downloads
- Retry logic

## Configuration

Priority order (highest to lowest):
1. Command-line flags
2. Environment variables (`GETBLOBZ_*`)
3. Config file (`getblobz.yaml`)
4. Defaults

## State Database Schema

```sql
CREATE TABLE sync_runs (
    id INTEGER PRIMARY KEY,
    started_at DATETIME,
    completed_at DATETIME,
    status TEXT,
    total_files INTEGER,
    downloaded_files INTEGER
);

CREATE TABLE blob_state (
    id INTEGER PRIMARY KEY,
    blob_name TEXT UNIQUE,
    etag TEXT,
    last_modified DATETIME,
    status TEXT,
    error_message TEXT
);
```

## Development

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Lint
golangci-lint run

# Build
go build -o getblobz main.go
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes and add tests
4. Submit a pull request

## Performance Considerations

- Use concurrent downloads (configurable worker pool)
- SQLite WAL mode for better concurrency
- Stream downloads (no in-memory buffering)
- Batch database operations
- Auto-throttling on resource-constrained devices
