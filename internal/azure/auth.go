// Package azure provides Azure Blob Storage client functionality and authentication.
package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/haepapa/getblobz/internal/config"
)

// CreateClient creates an Azure Blob Storage client based on the provided configuration.
// It supports multiple authentication methods: connection string, account key,
// managed identity, service principal, and Azure CLI credentials.
func CreateClient(cfg *config.AzureConfig) (*azblob.Client, error) {
	if cfg.ConnectionString != "" {
		return createClientFromConnectionString(cfg.ConnectionString)
	}

	if cfg.AccountName != "" {
		return createClientFromAccountName(cfg)
	}

	return nil, fmt.Errorf("no valid authentication method configured")
}

// createClientFromConnectionString creates a client using a connection string.
func createClientFromConnectionString(connectionString string) (*azblob.Client, error) {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client from connection string: %w", err)
	}
	return client, nil
}

// createClientFromAccountName creates a client using account name with various auth methods.
func createClientFromAccountName(cfg *config.AzureConfig) (*azblob.Client, error) {
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", cfg.AccountName)

	if cfg.AccountKey != "" {
		cred, err := azblob.NewSharedKeyCredential(cfg.AccountName, cfg.AccountKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create shared key credential: %w", err)
		}
		client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create client with shared key: %w", err)
		}
		return client, nil
	}

	if cfg.UseManagedIdentity {
		cred, err := azidentity.NewManagedIdentityCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create managed identity credential: %w", err)
		}
		client, err := azblob.NewClient(serviceURL, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create client with managed identity: %w", err)
		}
		return client, nil
	}

	if cfg.TenantID != "" && cfg.ClientID != "" && cfg.ClientSecret != "" {
		cred, err := azidentity.NewClientSecretCredential(
			cfg.TenantID,
			cfg.ClientID,
			cfg.ClientSecret,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create service principal credential: %w", err)
		}
		client, err := azblob.NewClient(serviceURL, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create client with service principal: %w", err)
		}
		return client, nil
	}

	if cfg.UseAzureCLI {
		cred, err := azidentity.NewAzureCLICredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure CLI credential: %w", err)
		}
		client, err := azblob.NewClient(serviceURL, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create client with Azure CLI: %w", err)
		}
		return client, nil
	}

	return nil, fmt.Errorf("no valid authentication method found for account name")
}
