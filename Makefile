.PHONY: help test test-unit test-integration test-e2e test-all coverage lint fmt vet clean build build-release build-all checksums docker-build docker-run prepare-release

# Default target
help:
	@echo "getblobz - Makefile targets:"
	@echo ""
	@echo "Testing:"
	@echo "  make test-unit         Run unit tests only (fast)"
	@echo "  make test-integration  Run integration tests with Azurite"
	@echo "  make test-e2e          Run end-to-end tests"
	@echo "  make test-all          Run all tests"
	@echo "  make coverage          Generate coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint              Run linters"
	@echo "  make fmt               Format code"
	@echo "  make vet               Run go vet"
	@echo "  make ci                Run all checks"
	@echo ""
	@echo "Build:"
	@echo "  make build             Build binary"
	@echo "  make build-release     Build with version info"
	@echo "  make build-all         Build for all platforms"
	@echo "  make docker-build      Build Docker image"
	@echo "  make prepare-release   Prepare release artifacts"
	@echo "  make clean             Clean build artifacts"
	@echo ""

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	@go test -v -short ./...

# Run integration tests
test-integration:
	@./scripts/test-integration.sh

# Run E2E tests
test-e2e:
	@./scripts/test-e2e.sh

# Run all tests
test-all: test-unit test-integration test-e2e

# Alias for test-unit
test: test-unit

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total

# Run golangci-lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  brew install golangci-lint (macOS)"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -w .

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Build binary
build:
	@echo "Building getblobz..."
	@go build -o getblobz main.go
	@echo "Built: getblobz"



# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f coverage.out coverage.html
	@rm -f getblobz getblobz-test
	@rm -rf dist/
	@echo "Clean complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Run all checks (CI equivalent)
ci: fmt vet lint test-unit
	@echo "✓ All CI checks passed"

# Release builds
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Build with version info
build-release:
	@echo "Building getblobz $(VERSION)..."
	@go build -ldflags="$(LDFLAGS)" -o getblobz main.go
	@echo "Built: getblobz ($(VERSION))"
	@./getblobz --version

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/getblobz-$(VERSION)-linux-amd64 main.go
	@GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/getblobz-$(VERSION)-linux-arm64 main.go
	@GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/getblobz-$(VERSION)-darwin-amd64 main.go
	@GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/getblobz-$(VERSION)-darwin-arm64 main.go
	@GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/getblobz-$(VERSION)-windows-amd64.exe main.go
	@echo "✓ Built binaries in dist/"
	@ls -lh dist/

# Generate checksums
checksums:
	@echo "Generating checksums..."
	@cd dist && for file in getblobz-*; do \
		sha256sum "$$file" > "$$file.sha256"; \
	done
	@echo "✓ Checksums generated"

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t getblobz:$(VERSION) .
	@docker tag getblobz:$(VERSION) getblobz:latest
	@echo "✓ Docker image built: getblobz:$(VERSION)"

# Docker run
docker-run:
	@docker run --rm getblobz:$(VERSION) --version

# Release preparation
prepare-release: clean build-all checksums
	@echo "✓ Release preparation complete"
	@echo "  Version: $(VERSION)"
	@echo "  Artifacts in dist/"
