#!/bin/bash
set -e

echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║              getblobz - Local Test Suite                         ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo ""

echo "→ Running unit tests..."
go test -v -short ./...

echo ""
echo "→ Running tests with coverage..."
go test -coverprofile=coverage.out ./...
echo ""
echo "Coverage summary:"
go tool cover -func=coverage.out | grep total

echo ""
echo "→ Running go vet..."
go vet ./...

echo ""
echo "→ Checking code formatting..."
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
    echo "❌ Files need formatting:"
    echo "$unformatted"
    exit 1
fi

echo ""
echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║              ✓ All local tests passed!                           ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
