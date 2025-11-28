// Package config manages configuration loading and validation for getblobz.
// It supports multiple configuration sources: files, environment variables, and command-line flags.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the complete application configuration.
type Config struct {
	Azure       AzureConfig       `mapstructure:"azure"`
	Sync        SyncConfig        `mapstructure:"sync"`
	Watch       WatchConfig       `mapstructure:"watch"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	State       StateConfig       `mapstructure:"state"`
	Performance PerformanceConfig `mapstructure:"performance"`
}

// AzureConfig contains Azure Storage authentication and connection settings.
type AzureConfig struct {
	// ConnectionString is the Azure Storage connection string.
	ConnectionString string `mapstructure:"connection_string"`
	// AccountName is the Azure Storage account name.
	AccountName string `mapstructure:"account_name"`
	// AccountKey is the Azure Storage account key.
	AccountKey string `mapstructure:"account_key"`
	// UseManagedIdentity enables Azure Managed Identity authentication.
	UseManagedIdentity bool `mapstructure:"use_managed_identity"`
	// TenantID is the Azure AD tenant ID for service principal authentication.
	TenantID string `mapstructure:"tenant_id"`
	// ClientID is the Azure AD client ID for service principal authentication.
	ClientID string `mapstructure:"client_id"`
	// ClientSecret is the Azure AD client secret for service principal authentication.
	ClientSecret string `mapstructure:"client_secret"`
	// UseAzureCLI enables Azure CLI credential authentication.
	UseAzureCLI bool `mapstructure:"use_azure_cli"`
}

// SyncConfig contains synchronisation operation settings.
type SyncConfig struct {
	// Container is the Azure Blob Storage container name.
	Container string `mapstructure:"container"`
	// OutputPath is the local directory where files will be downloaded.
	OutputPath string `mapstructure:"output_path"`
	// Prefix filters blobs to only those starting with this prefix.
	Prefix string `mapstructure:"prefix"`
	// Workers specifies the number of concurrent download workers.
	Workers int `mapstructure:"workers"`
	// BatchSize is the number of blobs to list per API call.
	BatchSize int `mapstructure:"batch_size"`
	// SkipExisting skips downloading files that already exist locally.
	SkipExisting bool `mapstructure:"skip_existing"`
	// VerifyChecksums enables MD5 checksum verification after download.
	VerifyChecksums bool `mapstructure:"verify_checksums"`
	// ForceResync forces re-download of all files ignoring state.
	ForceResync bool `mapstructure:"force_resync"`
	// DiskWarnPercent is the filesystem usage percent at which a warning is logged.
	DiskWarnPercent int `mapstructure:"disk_warn_percent"`
	// DiskStopPercent is the filesystem usage percent at which downloads stop.
	DiskStopPercent int `mapstructure:"disk_stop_percent"`
	// FolderOrganization contains settings for organizing files into folders.
	FolderOrganization FolderOrganizationConfig `mapstructure:"folder_organization"`
}

// FolderOrganizationConfig contains settings for organizing downloaded files into folders.
type FolderOrganizationConfig struct {
	// Enabled enables automatic folder organization.
	Enabled bool `mapstructure:"enabled"`
	// MaxFilesPerFolder is the maximum number of files per folder.
	MaxFilesPerFolder int `mapstructure:"max_files_per_folder"`
	// Strategy determines the folder organization strategy (partition_key, date, sequential).
	Strategy string `mapstructure:"strategy"`
	// PartitionDepth is the depth of partition key hashing (for partition_key strategy).
	PartitionDepth int `mapstructure:"partition_depth"`
}

// WatchConfig contains continuous sync monitoring settings.
type WatchConfig struct {
	// Enabled enables continuous watch mode.
	Enabled bool `mapstructure:"enabled"`
	// Interval is the duration between sync runs in watch mode.
	Interval time.Duration `mapstructure:"interval"`
}

// LoggingConfig contains logging configuration.
type LoggingConfig struct {
	// Level specifies the minimum log level (debug, info, warn, error).
	Level string `mapstructure:"level"`
	// Format specifies the log output format (text, json).
	Format string `mapstructure:"format"`
}

// StateConfig contains state database configuration.
type StateConfig struct {
	// Database is the path to the SQLite state database file.
	Database string `mapstructure:"database"`
}

// PerformanceConfig contains performance tuning and resource limit settings.
type PerformanceConfig struct {
	// MaxMemoryMB limits maximum memory usage in megabytes (0 = auto-detect).
	MaxMemoryMB int `mapstructure:"max_memory_mb"`
	// MaxCPUPercent limits maximum CPU utilisation percentage.
	MaxCPUPercent int `mapstructure:"max_cpu_percent"`
	// AutoThrottle enables automatic throttling based on system load.
	AutoThrottle bool `mapstructure:"auto_throttle"`
	// ThrottleThreshold is the system load threshold for throttling.
	ThrottleThreshold float64 `mapstructure:"throttle_threshold"`
	// BandwidthLimit limits network bandwidth (e.g., "10M", "100K").
	BandwidthLimit string `mapstructure:"bandwidth_limit"`
	// DiskBufferMB is the disk write buffer size in megabytes.
	DiskBufferMB int `mapstructure:"disk_buffer_mb"`
}

// Default returns a Config with sensible default values.
func Default() *Config {
	return &Config{
		Sync: SyncConfig{
			OutputPath:      "./data",
			Workers:         10,
			BatchSize:       5000,
			SkipExisting:    true,
			VerifyChecksums: true,
			DiskWarnPercent: 80,
			DiskStopPercent: 90,
			FolderOrganization: FolderOrganizationConfig{
				Enabled:           false,
				MaxFilesPerFolder: 10000,
				Strategy:          "sequential",
				PartitionDepth:    2,
			},
		},
		Watch: WatchConfig{
			Enabled:  false,
			Interval: 5 * time.Minute,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
		State: StateConfig{
			Database: "./.sync-state.db",
		},
		Performance: PerformanceConfig{
			MaxMemoryMB:       0,
			MaxCPUPercent:     80,
			AutoThrottle:      false,
			ThrottleThreshold: 0.8,
			DiskBufferMB:      32,
		},
	}
}

// Validate checks if the configuration is valid and returns an error if not.
func (c *Config) Validate() error {
	if c.Sync.Container == "" {
		return fmt.Errorf("container name is required")
	}

	if c.Azure.ConnectionString == "" && c.Azure.AccountName == "" {
		return fmt.Errorf("either connection string or account name must be provided")
	}

	if c.Azure.AccountName != "" && c.Azure.ConnectionString == "" {
		hasAuth := c.Azure.AccountKey != "" ||
			c.Azure.UseManagedIdentity ||
			(c.Azure.TenantID != "" && c.Azure.ClientID != "" && c.Azure.ClientSecret != "") ||
			c.Azure.UseAzureCLI

		if !hasAuth {
			return fmt.Errorf("authentication method required when using account name")
		}
	}

	if c.Sync.Workers < 1 || c.Sync.Workers > 100 {
		return fmt.Errorf("workers must be between 1 and 100")
	}

	if c.Sync.BatchSize < 1 || c.Sync.BatchSize > 10000 {
		return fmt.Errorf("batch size must be between 1 and 10000")
	}

	if c.Sync.DiskWarnPercent < 1 || c.Sync.DiskWarnPercent > 99 {
		return fmt.Errorf("disk warn percent must be between 1 and 99")
	}
	if c.Sync.DiskStopPercent < 1 || c.Sync.DiskStopPercent > 99 {
		return fmt.Errorf("disk stop percent must be between 1 and 99")
	}
	if c.Sync.DiskWarnPercent >= c.Sync.DiskStopPercent {
		return fmt.Errorf("disk warn percent must be less than disk stop percent")
	}

	if c.Performance.MaxCPUPercent < 1 || c.Performance.MaxCPUPercent > 100 {
		return fmt.Errorf("max CPU percent must be between 1 and 100")
	}

	if c.Performance.ThrottleThreshold < 0.1 || c.Performance.ThrottleThreshold > 1.0 {
		return fmt.Errorf("throttle threshold must be between 0.1 and 1.0")
	}

	if c.Sync.FolderOrganization.Enabled {
		if c.Sync.FolderOrganization.MaxFilesPerFolder < 100 || c.Sync.FolderOrganization.MaxFilesPerFolder > 100000 {
			return fmt.Errorf("max files per folder must be between 100 and 100000")
		}

		validStrategies := map[string]bool{
			"sequential":    true,
			"partition_key": true,
			"date":          true,
		}
		if !validStrategies[c.Sync.FolderOrganization.Strategy] {
			return fmt.Errorf("invalid folder organization strategy: must be sequential, partition_key, or date")
		}

		if c.Sync.FolderOrganization.PartitionDepth < 1 || c.Sync.FolderOrganization.PartitionDepth > 4 {
			return fmt.Errorf("partition depth must be between 1 and 4")
		}
	}

	return nil
}

// GetConfigPath returns the configuration file path based on priority:
// 1. Explicit path if provided
// 2. Current directory (./getblobz.yaml or ./getblobz.yml)
// 3. User config directory (~/.config/getblobz/config.yaml on Unix)
func GetConfigPath(explicitPath string) string {
	if explicitPath != "" {
		return explicitPath
	}

	if _, err := os.Stat("./getblobz.yaml"); err == nil {
		return "./getblobz.yaml"
	}

	if _, err := os.Stat("./getblobz.yml"); err == nil {
		return "./getblobz.yml"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configDir := filepath.Join(homeDir, ".config", "getblobz", "config.yaml")
	if _, err := os.Stat(configDir); err == nil {
		return configDir
	}

	return ""
}
