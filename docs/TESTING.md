# Testing Guide

## Running Tests

```bash
# All tests
go test ./...

# Unit tests only
go test -short ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Specific package
go test -v ./internal/config
```

## Test Structure

```
getblobz/
├── internal/
│   ├── config/
│   │   └── config_test.go
│   ├── storage/
│   │   └── db_test.go
│   └── sync/
│       └── syncer_test.go
└── test/
    ├── integration/
    └── e2e/
```

## Integration Tests

Requires Azurite (Azure Storage Emulator):

```bash
# Install Azurite
npm install -g azurite

# Start Azurite
azurite --silent &

# Run integration tests
go test -tags=integration ./test/integration/...
```

## End-to-End Tests

```bash
go test -tags=e2e ./test/e2e/...
```

## Writing Tests

### Unit Test Example

```go
func TestConfig_Validate(t *testing.T) {
    cfg := &Config{
        Azure: AzureConfig{ConnectionString: "..."},
        Sync:  SyncConfig{Container: "test"},
    }
    
    err := cfg.Validate()
    if err != nil {
        t.Errorf("expected valid config, got error: %v", err)
    }
}
```

### Table-Driven Test

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid input", "test", false},
        {"empty input", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, want error %v", err, tt.wantErr)
            }
        })
    }
}
```

## CI/CD

Tests run automatically on GitHub Actions for:
- Pull requests
- Merges to main/test/prod branches

## Coverage Goals

- Overall: 80%+
- Critical packages (config, storage): 85%+
