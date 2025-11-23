# Building getblobz

This document provides detailed build instructions for getblobz.

## Prerequisites

- Go 1.21 or higher
- Git
- C compiler (for SQLite CGO bindings)
  - macOS: Xcode Command Line Tools
  - Linux: gcc or clang
  - Windows: MinGW-w64 or TDM-GCC

## Quick Build

```bash
# Clone the repository
git clone https://github.com/haepapa/getblobz.git
cd getblobz

# Build
go build -o getblobz main.go

# Run
./getblobz --help
```

## Development Build

For development with debugging symbols:

```bash
go build -v -o getblobz main.go
```

## Production Build

Optimized build with reduced binary size:

```bash
go build -ldflags="-s -w" -o getblobz main.go
```

Flags explained:
- `-s`: Omit symbol table
- `-w`: Omit DWARF debugging information

## Cross-Compilation

### Linux (amd64)

```bash
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o getblobz-linux main.go
```

### Linux (arm64) - Raspberry Pi

```bash
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o getblobz-arm64 main.go
```

### macOS (amd64)

```bash
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o getblobz-darwin-amd64 main.go
```

### macOS (arm64) - Apple Silicon

```bash
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o getblobz-darwin-arm64 main.go
```

### Windows (amd64)

```bash
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o getblobz-windows.exe main.go
```

**Note:** Cross-compilation with CGO requires the appropriate C cross-compiler toolchain for the target platform.

## Docker Build

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -ldflags="-s -w" -o getblobz main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/getblobz .

ENTRYPOINT ["./getblobz"]
```

Build the Docker image:

```bash
docker build -t getblobz:latest .
```

Run:

```bash
docker run --rm -v $(pwd)/downloads:/downloads \
  getblobz:latest sync \
  --container mycontainer \
  --connection-string "..." \
  --output-path /downloads
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Package Tests

```bash
go test -v ./internal/config
go test -v ./internal/storage
go test -v ./internal/sync
```

## Code Quality

### Format Code

```bash
go fmt ./...
```

### Lint Code

Install golangci-lint:

```bash
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Windows
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

Run linter:

```bash
golangci-lint run
```

### Check for Issues

```bash
go vet ./...
```

### Static Analysis

```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

## Dependency Management

### Update Dependencies

```bash
go get -u ./...
go mod tidy
```

### Verify Dependencies

```bash
go mod verify
```

### List Dependencies

```bash
go list -m all
```

### Vendor Dependencies

To vendor dependencies for offline builds:

```bash
go mod vendor
```

Then build with:

```bash
go build -mod=vendor -o getblobz main.go
```

## Benchmarking

Run benchmarks:

```bash
go test -bench=. -benchmem ./...
```

## Installation

### Install to GOPATH

```bash
go install
```

### Install to System

```bash
# Build
go build -o getblobz main.go

# Install (macOS/Linux)
sudo mv getblobz /usr/local/bin/

# Install (Windows - run as Administrator)
move getblobz.exe C:\Windows\System32\
```

## Troubleshooting

### CGO Issues

If you encounter CGO-related errors:

```bash
# Disable CGO (will use pure Go SQLite implementation)
CGO_ENABLED=0 go build -o getblobz main.go
```

Note: This may impact performance.

### Module Issues

Clear module cache:

```bash
go clean -modcache
go mod download
```

### Build Errors

Ensure you have the correct Go version:

```bash
go version  # Should be 1.21 or higher
```

Update go.mod if needed:

```bash
go mod edit -go=1.21
go mod tidy
```

## CI/CD Integration

### GitHub Actions

Create `.github/workflows/build.yml`:

```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build
        run: go build -v -o getblobz main.go
      
      - name: Test
        run: go test -v ./...
      
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: getblobz
          path: getblobz
```

## Release Process

1. Update version in `cmd/root.go`
2. Tag the release:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. Build binaries for all platforms
4. Create GitHub release with binaries

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Modules](https://golang.org/ref/mod)
- [CGO Documentation](https://golang.org/cmd/cgo/)
