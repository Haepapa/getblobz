#!/bin/bash
set -e

echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║         getblobz - Integration Test Suite                        ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo ""

if ! command -v azurite &> /dev/null; then
    echo "❌ Azurite is not installed"
    echo "Install with: npm install -g azurite"
    exit 1
fi

export AZURE_STORAGE_ALLOW_INSECURE_CONNECTION=true

echo "→ Starting Azurite (Azure Storage Emulator)..."
AZURITE_DIR=$(mktemp -d)
azurite --silent --location "$AZURITE_DIR" --skipApiVersionCheck --oauth basic &
AZURITE_PID=$!

trap "echo '→ Stopping Azurite...'; kill $AZURITE_PID 2>/dev/null || true; sleep 1; rm -rf $AZURITE_DIR 2>/dev/null || true" EXIT

echo "→ Waiting for Azurite to be ready..."
sleep 3

echo ""
echo "→ Running integration tests..."
CGO_ENABLED=0 go test -v -tags=integration ./test/integration/...

echo ""
echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║           ✓ Integration tests passed!                            ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
