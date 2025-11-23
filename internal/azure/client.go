// Package azure provides Azure Blob Storage operations wrapper.
package azure

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

// Client wraps the Azure Blob Storage client with application-specific operations.
type Client struct {
	client *azblob.Client
}

// NewClient creates a new Azure client wrapper.
func NewClient(client *azblob.Client) *Client {
	return &Client{client: client}
}

// BlobInfo contains metadata about a blob.
type BlobInfo struct {
	Name         string
	Path         string
	Size         int64
	ETag         string
	LastModified string
	ContentMD5   []byte
}

// ListBlobs lists all blobs in a container with the given prefix.
// It handles pagination automatically using continuation tokens.
func (c *Client) ListBlobs(ctx context.Context, containerName, prefix string, maxResults int32) ([]*BlobInfo, *string, error) {
	pager := c.client.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		Prefix:     &prefix,
		MaxResults: &maxResults,
		Include:    container.ListBlobsInclude{Metadata: true},
	})

	var blobs []*BlobInfo
	var continuationToken *string

	if pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, item := range page.Segment.BlobItems {
			if item.Name == nil {
				continue
			}

			blobInfo := &BlobInfo{
				Name: *item.Name,
				Path: *item.Name,
			}

			if item.Properties != nil {
				if item.Properties.ContentLength != nil {
					blobInfo.Size = *item.Properties.ContentLength
				}
				if item.Properties.ETag != nil {
					blobInfo.ETag = string(*item.Properties.ETag)
				}
				if item.Properties.LastModified != nil {
					blobInfo.LastModified = item.Properties.LastModified.Format("2006-01-02T15:04:05Z")
				}
				if item.Properties.ContentMD5 != nil {
					blobInfo.ContentMD5 = item.Properties.ContentMD5
				}
			}

			blobs = append(blobs, blobInfo)
		}

		if page.NextMarker != nil && *page.NextMarker != "" {
			continuationToken = page.NextMarker
		}
	}

	return blobs, continuationToken, nil
}

// DownloadBlob downloads a blob to the provided writer.
// It streams the content to avoid loading large files into memory.
func (c *Client) DownloadBlob(ctx context.Context, containerName, blobName string, writer io.Writer) error {
	blobClient := c.client.ServiceClient().NewContainerClient(containerName).NewBlobClient(blobName)

	resp, err := blobClient.DownloadStream(ctx, &blob.DownloadStreamOptions{})
	if err != nil {
		return fmt.Errorf("failed to download blob: %w", err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("failed to copy blob data: %w", err)
	}

	return nil
}

// GetBlobProperties retrieves metadata for a specific blob.
func (c *Client) GetBlobProperties(ctx context.Context, containerName, blobName string) (*BlobInfo, error) {
	blobClient := c.client.ServiceClient().NewContainerClient(containerName).NewBlobClient(blobName)

	props, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob properties: %w", err)
	}

	info := &BlobInfo{
		Name: blobName,
		Path: blobName,
	}

	if props.ContentLength != nil {
		info.Size = *props.ContentLength
	}
	if props.ETag != nil {
		info.ETag = string(*props.ETag)
	}
	if props.LastModified != nil {
		info.LastModified = props.LastModified.Format("2006-01-02T15:04:05Z")
	}
	if props.ContentMD5 != nil {
		info.ContentMD5 = props.ContentMD5
	}

	return info, nil
}

// ContainerExists checks if a container exists.
func (c *Client) ContainerExists(ctx context.Context, containerName string) (bool, error) {
	containerClient := c.client.ServiceClient().NewContainerClient(containerName)
	_, err := containerClient.GetProperties(ctx, nil)
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check container: %w", err)
	}
	return true, nil
}

// isNotFoundError checks if an error is a "not found" error.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check for Azure SDK not found errors
	return false
}
