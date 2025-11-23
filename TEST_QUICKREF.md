# getblobz Testing Quick Reference

## Essential Commands

### Local Testing
```bash
# Run unit tests (fast)
make test-unit

# Run with coverage
make coverage

# Run all quality checks
make ci

# Format and vet code
make fmt
make vet
```

### Integration Testing
```bash
# Start Azurite manually
./scripts/start-azurite.sh

# Run integration tests
make test-integration

# Run E2E tests
make test-e2e
```

### All Tests
```bash
make test-all
```

## Azurite Connection String

For local testing with Azurite emulator:

```
DefaultEndpointsProtocol=http;
AccountName=devstoreaccount1;
AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;
BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;
```

## Test File Structure

```
Package_test.go structure:

1. Test setup helpers
2. Test fixtures
3. Unit tests (TestFunctionName_Scenario_Expected)
4. Table-driven tests
5. Cleanup functions
```

## Common Test Patterns

### Table-Driven Test
```go
tests := []struct {
    name    string
    input   Type
    want    Type
    wantErr bool
}{
    {"description", input, expected, false},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test logic
    })
}
```

### Setup & Teardown
```go
func TestExample(t *testing.T) {
    // Setup
    resource := setupResource(t)
    defer cleanup(resource)
    
    // Test
    // ...
}
```

### Using t.Cleanup
```go
func TestWithCleanup(t *testing.T) {
    resource := createResource()
    t.Cleanup(func() {
        resource.Close()
    })
    // Test continues
}
```

## Build Tags

```go
// Unit tests (default)
// No build tag needed

// Integration tests
// +build integration

// E2E tests  
// +build e2e

// Load tests
// +build load
```

## Running Specific Tests

```bash
# Run tests in one package
go test ./internal/config

# Run specific test
go test -run TestValidate ./internal/config

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...

# Skip slow tests
go test -short ./...
```

## Coverage Commands

```bash
# Generate coverage
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out

# Check total coverage
go tool cover -func=coverage.out | grep total
```

## Debugging Tests

```bash
# Run test with detailed output
go test -v -run TestName ./package

# Print test with logging
go test -v ./package 2>&1 | tee test.log

# Run single test function
go test -v -run '^TestFunctionName$' ./package
```

## Mock Generation

```bash
# Using mockgen
mockgen -destination=mocks/mock_type.go \
  -package=mocks \
  github.com/haepapa/getblobz/internal/package Interface

# Using mockery
mockery --name=Interface --output=mocks
```

## CI/CD

Tests run automatically on:
- Push to `main` or `develop`
- Pull requests to `main` or `develop`

View results at:
- GitHub Actions tab
- PR status checks

## Coverage Thresholds

| Package | Target |
|---------|--------|
| internal/config | 90% |
| internal/storage | 85% |
| internal/azure | 75% |
| internal/sync | 80% |
| pkg/logger | 85% |
| **Overall** | **80%** |

## Test Dependencies

### Required
- Go 1.21+
- Azurite (for integration tests)

### Optional
- golangci-lint (for linting)
- gotestsum (for better output)
- gocov (for coverage tools)

## Installation

```bash
# Azurite
npm install -g azurite

# Test tools
go install github.com/golang/mock/mockgen@latest
go install gotest.tools/gotestsum@latest

# Linter
brew install golangci-lint  # macOS
# or
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Troubleshooting

### "Azurite not found"
```bash
npm install -g azurite
```

### "Tests hang"
- Check for goroutine leaks
- Use `go test -timeout=30s`

### "Database locked"
- Ensure proper cleanup
- Check for unclosed database connections

### "Flaky tests"
- Remove time dependencies
- Use proper synchronization
- Avoid global state

## Best Practices

1. ✅ Write tests first (TDD)
2. ✅ Keep tests fast
3. ✅ Use descriptive test names
4. ✅ Test edge cases
5. ✅ Mock external dependencies
6. ✅ Clean up resources
7. ✅ Use table-driven tests
8. ✅ Maintain high coverage
9. ✅ Run tests before commit
10. ✅ Keep tests simple

## Resources

- **Full Documentation**: [TESTERNOTES.md](TESTERNOTES.md)
- **Build Guide**: [BUILD.md](BUILD.md)
- **Developer Notes**: [DEVELOPERNOTES.md](DEVELOPERNOTES.md)
- **Go Testing**: https://golang.org/pkg/testing/
- **Testify**: https://github.com/stretchr/testify
- **Azurite**: https://github.com/Azure/Azurite

---

**Quick Help**: For detailed examples and complete test implementations, see TESTERNOTES.md
