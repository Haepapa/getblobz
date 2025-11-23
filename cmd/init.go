// Package cmd provides the init command for configuration file generation.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// initCmd represents the init command.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a configuration file template",
	Long: `Init creates a configuration file template with example values.

The generated file can be customised and used with the --config flag
or placed in one of the auto-discovery locations.

Examples:
  # Generate config in current directory
  getblobz init

  # Generate config at specific path
  getblobz init --config /path/to/config.yaml`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := cfgFile
	if configPath == "" {
		configPath = "./getblobz.yaml"
	}

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	template := `# getblobz configuration file
# Documentation: https://github.com/haepapa/getblobz

azure:
  # Option 1: Connection string
  connection_string: "DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;EndpointSuffix=core.windows.net"
  
  # Option 2: Account name + key
  # account_name: "mystorageaccount"
  # account_key: "key=="
  
  # Option 3: Managed Identity (Azure VM/Container)
  # account_name: "mystorageaccount"
  # use_managed_identity: true
  
  # Option 4: Service Principal
  # account_name: "mystorageaccount"
  # tenant_id: "tenant-id"
  # client_id: "client-id"
  # client_secret: "client-secret"
  
  # Option 5: Azure CLI
  # account_name: "mystorageaccount"
  # use_azure_cli: true

sync:
  container: "mycontainer"
  output_path: "./downloads"
  prefix: ""                  # Optional: filter blobs by prefix
  workers: 10                 # Concurrent download workers
  batch_size: 5000            # Blobs per listing batch
  skip_existing: true         # Skip already downloaded files
  verify_checksums: true      # Verify MD5 after download

watch:
  enabled: false              # Continuous monitoring mode
  interval: "5m"              # Check interval (e.g., "5m", "1h")

logging:
  level: "info"               # debug, info, warn, error
  format: "text"              # text, json

state:
  database: "./.sync-state.db"  # SQLite state database path

performance:
  max_memory_mb: 0            # 0 = auto-detect
  max_cpu_percent: 80         # Maximum CPU utilization
  auto_throttle: false        # Enable automatic throttling
  throttle_threshold: 0.8     # System load threshold for throttling
  bandwidth_limit: ""         # e.g., "50M" for 50 MB/s
  disk_buffer_mb: 32          # Disk write buffer size
`

	if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Configuration file created at: %s\n", configPath)
	fmt.Println("\nEdit the file to add your credentials and settings, then run:")
	fmt.Println("  getblobz sync")

	return nil
}
