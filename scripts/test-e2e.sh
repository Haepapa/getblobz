#!/bin/bash
set -e

echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║            getblobz - End-to-End Test Suite                      ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo ""

if ! command -v azurite &> /dev/null; then
    echo "❌ Azurite is not installed"
    echo "Install with: npm install -g azurite"
    exit 1
fi

echo "→ Building application..."
go build -o getblobz-test main.go

trap "rm -f getblobz-test" EXIT

echo "→ Starting Azurite..."
AZURITE_DIR=$(mktemp -d)
azurite --silent --location "$AZURITE_DIR" &
AZURITE_PID=$!

trap "echo '→ Stopping Azurite...'; kill $AZURITE_PID 2>/dev/null || true; rm -rf $AZURITE_DIR; rm -f getblobz-test" EXIT

sleep 3

echo ""
echo "→ Running E2E tests..."
go test -v -tags=e2e -timeout=5m ./test/e2e/...

echo ""
echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║              ✓ E2E tests passed!                                 ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
