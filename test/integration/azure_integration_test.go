//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/haepapa/getblobz/internal/azure"
)

// getAzuriteConnString returns the Azurite connection string, defaulting to local emulator.
func getAzuriteConnString() string {
	if v := os.Getenv("AZURITE_CONNECTION_STRING"); v != "" {
		return v
	}
	return "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;"
}

func TestAzureClient_ListAndDownload_WithAzurite(t *testing.T) {
	ctx := context.Background()

	// Use shared key credential with HTTP endpoint
	accountName := "devstoreaccount1"
	accountKey := "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
	serviceURL := "http://127.0.0.1:10000/devstoreaccount1"

	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	// Create client with InsecureAllowCredentialWithHTTP enabled for Azurite
	clientOpts := &azblob.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			InsecureAllowCredentialWithHTTP: true,
		},
	}
	sdkClient, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, clientOpts)
	if err != nil {
		t.Fatalf("failed to create azblob client: %v", err)
	}

	// Wrap in our azure.Client
	c := azure.NewClient(sdkClient)

	containerName := "it-container-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	blobName := "hello.txt"
	blobContent := []byte("hello integration")

	// Create container
	contClient := sdkClient.ServiceClient().NewContainerClient(containerName)
	_, err = contClient.Create(ctx, nil)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	defer func() { _, _ = contClient.Delete(ctx, nil) }()

	// Upload blob
	bClient := contClient.NewBlockBlobClient(blobName)
	_, err = bClient.UploadBuffer(ctx, blobContent, nil)
	if err != nil {
		t.Fatalf("failed to upload blob: %v", err)
	}

	// List via wrapper
	blobs, _, err := c.ListBlobs(ctx, containerName, "", 100)
	if err != nil {
		t.Fatalf("ListBlobs error: %v", err)
	}
	if len(blobs) != 1 || blobs[0].Name != blobName {
		t.Fatalf("expected one blob %q, got: %+v", blobName, blobs)
	}

	// Download via wrapper and verify content
	var got bytes.Buffer
	if err := c.DownloadBlob(ctx, containerName, blobName, &got); err != nil {
		t.Fatalf("DownloadBlob error: %v", err)
	}
	if got.String() != string(blobContent) {
		t.Fatalf("downloaded content mismatch: got %q, want %q", got.String(), string(blobContent))
	}
}
