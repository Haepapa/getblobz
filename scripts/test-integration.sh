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

echo "→ Starting Azurite (Azure Storage Emulator)..."
AZURITE_DIR=$(mktemp -d)
azurite --silent --location "$AZURITE_DIR" &
AZURITE_PID=$!

trap "echo '→ Stopping Azurite...'; kill $AZURITE_PID 2>/dev/null || true; rm -rf $AZURITE_DIR" EXIT

echo "→ Waiting for Azurite to be ready..."
sleep 3

echo ""
echo "→ Running integration tests..."
go test -v -tags=integration ./test/integration/...

echo ""
echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║           ✓ Integration tests passed!                            ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
