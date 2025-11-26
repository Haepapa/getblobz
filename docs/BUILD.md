# Building getblobz

## Prerequisites

- Go 1.21+
- C compiler (for SQLite)
  - macOS: Xcode Command Line Tools
  - Linux: gcc
  - Windows: MinGW-w64

## Quick Build

```bash
git clone https://github.com/haepapa/getblobz.git
cd getblobz
go build -o getblobz main.go
```

## Production Build

```bash
go build -ldflags="-s -w" -o getblobz main.go
```

## Cross-Compilation

```bash
# Linux
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o getblobz-linux main.go

# macOS (Apple Silicon)
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o getblobz-darwin-arm64 main.go

# Windows
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o getblobz.exe main.go
```

## Testing

```bash
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Installation

```bash
# System-wide
sudo mv getblobz /usr/local/bin/

# Or to GOPATH
go install
```
