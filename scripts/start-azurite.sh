#!/bin/bash

echo "Starting Azurite (Azure Storage Emulator)..."
echo ""
echo "Connection String:"
echo "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;"
echo ""
echo "Press Ctrl+C to stop"
echo ""

azurite --location /tmp/azurite
