// Package cmd provides the sync command implementation.
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/haepapa/getblobz/internal/azure"
	"github.com/haepapa/getblobz/internal/storage"
	"github.com/haepapa/getblobz/internal/sync"
	"github.com/haepapa/getblobz/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// syncCmd represents the sync command.
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronise blobs from Azure Storage to local filesystem",
	Long: `Sync downloads blobs from an Azure Blob Storage container to a local directory.

It supports incremental sync, continuous monitoring, and handles interruptions gracefully.
State is tracked in a SQLite database to enable resumable operations.

Examples:
  # Initial sync
  getblobz sync --container mycontainer --connection-string "..."

  # Continuous sync with watch mode
  getblobz sync --container mycontainer --connection-string "..." --watch

  # Sync with prefix filter
  getblobz sync --container mycontainer --connection-string "..." --prefix "data/2024/"`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().String("container", "", "Azure container name (required)")
	syncCmd.Flags().String("output-path", "./data", "local destination path")
	syncCmd.Flags().String("connection-string", "", "Azure Storage connection string")
	syncCmd.Flags().String("account-name", "", "Storage account name")
	syncCmd.Flags().String("account-key", "", "Storage account key")
	syncCmd.Flags().Bool("use-managed-identity", false, "use Azure Managed Identity")
	syncCmd.Flags().String("tenant-id", "", "Azure AD tenant ID")
	syncCmd.Flags().String("client-id", "", "Azure AD client ID")
	syncCmd.Flags().String("client-secret", "", "Azure AD client secret")
	syncCmd.Flags().Bool("use-azure-cli", false, "use Azure CLI credentials")
	syncCmd.Flags().String("prefix", "", "only sync blobs with this prefix")
	syncCmd.Flags().Int("workers", 10, "number of concurrent download workers")
	syncCmd.Flags().Int("batch-size", 5000, "number of blobs to list per batch")
	syncCmd.Flags().Bool("watch", false, "continuously watch for new files")
	syncCmd.Flags().Duration("watch-interval", 5*time.Minute, "interval between checks in watch mode")
	syncCmd.Flags().String("state-db", "./.sync-state.db", "path to state database")
	syncCmd.Flags().Bool("force-resync", false, "ignore state and re-download all files")
	syncCmd.Flags().Bool("skip-existing", true, "skip files that already exist locally")
	syncCmd.Flags().Bool("verify-checksums", true, "verify MD5 checksums after download")
	syncCmd.Flags().Int("disk-warn-percent", 80, "filesystem usage percent to warn at (1-99)")
	syncCmd.Flags().Int("disk-stop-percent", 90, "filesystem usage percent to stop at (1-99)")
	syncCmd.Flags().Bool("organize-folders", false, "enable folder organization")
	syncCmd.Flags().Int("max-files-per-folder", 10000, "maximum files per folder")
	syncCmd.Flags().String("folder-strategy", "sequential", "folder organization strategy (sequential, partition_key, date)")
	syncCmd.Flags().Int("partition-depth", 2, "partition depth for partition_key strategy")

	if err := syncCmd.MarkFlagRequired("container"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to mark required flag: %v\n", err)
	}

	if err := viper.BindPFlag("azure.connection_string", syncCmd.Flags().Lookup("connection-string")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind connection-string: %v\n", err)
	}
	if err := viper.BindPFlag("azure.account_name", syncCmd.Flags().Lookup("account-name")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind account-name: %v\n", err)
	}
	if err := viper.BindPFlag("azure.account_key", syncCmd.Flags().Lookup("account-key")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind account-key: %v\n", err)
	}
	if err := viper.BindPFlag("azure.use_managed_identity", syncCmd.Flags().Lookup("use-managed-identity")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind use-managed-identity: %v\n", err)
	}
	if err := viper.BindPFlag("azure.tenant_id", syncCmd.Flags().Lookup("tenant-id")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind tenant-id: %v\n", err)
	}
	if err := viper.BindPFlag("azure.client_id", syncCmd.Flags().Lookup("client-id")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind client-id: %v\n", err)
	}
	if err := viper.BindPFlag("azure.client_secret", syncCmd.Flags().Lookup("client-secret")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind client-secret: %v\n", err)
	}
	if err := viper.BindPFlag("azure.use_azure_cli", syncCmd.Flags().Lookup("use-azure-cli")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind use-azure-cli: %v\n", err)
	}
	if err := viper.BindPFlag("sync.container", syncCmd.Flags().Lookup("container")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind container: %v\n", err)
	}
	if err := viper.BindPFlag("sync.output_path", syncCmd.Flags().Lookup("output-path")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind output-path: %v\n", err)
	}
	if err := viper.BindPFlag("sync.prefix", syncCmd.Flags().Lookup("prefix")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind prefix: %v\n", err)
	}
	if err := viper.BindPFlag("sync.workers", syncCmd.Flags().Lookup("workers")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind workers: %v\n", err)
	}
	if err := viper.BindPFlag("sync.batch_size", syncCmd.Flags().Lookup("batch-size")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind batch-size: %v\n", err)
	}
	if err := viper.BindPFlag("sync.skip_existing", syncCmd.Flags().Lookup("skip-existing")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind skip-existing: %v\n", err)
	}
	if err := viper.BindPFlag("sync.verify_checksums", syncCmd.Flags().Lookup("verify-checksums")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind verify-checksums: %v\n", err)
	}
	if err := viper.BindPFlag("sync.force_resync", syncCmd.Flags().Lookup("force-resync")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind force-resync: %v\n", err)
	}
	if err := viper.BindPFlag("sync.disk_warn_percent", syncCmd.Flags().Lookup("disk-warn-percent")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind disk-warn-percent: %v\n", err)
	}
	if err := viper.BindPFlag("sync.disk_stop_percent", syncCmd.Flags().Lookup("disk-stop-percent")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind disk-stop-percent: %v\n", err)
	}
	if err := viper.BindPFlag("sync.folder_organization.enabled", syncCmd.Flags().Lookup("organize-folders")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind organize-folders: %v\n", err)
	}
	if err := viper.BindPFlag("sync.folder_organization.max_files_per_folder", syncCmd.Flags().Lookup("max-files-per-folder")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind max-files-per-folder: %v\n", err)
	}
	if err := viper.BindPFlag("sync.folder_organization.strategy", syncCmd.Flags().Lookup("folder-strategy")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind folder-strategy: %v\n", err)
	}
	if err := viper.BindPFlag("sync.folder_organization.partition_depth", syncCmd.Flags().Lookup("partition-depth")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind partition-depth: %v\n", err)
	}
	if err := viper.BindPFlag("watch.enabled", syncCmd.Flags().Lookup("watch")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind watch: %v\n", err)
	}
	if err := viper.BindPFlag("watch.interval", syncCmd.Flags().Lookup("watch-interval")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind watch-interval: %v\n", err)
	}
	if err := viper.BindPFlag("state.database", syncCmd.Flags().Lookup("state-db")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind state-db: %v\n", err)
	}
}

func runSync(cmd *cobra.Command, args []string) error {
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	log, err := logger.New(logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer func() { _ = log.Close() }()

	db, err := storage.Open(cfg.State.Database)
	if err != nil {
		return fmt.Errorf("failed to open state database: %w", err)
	}
	defer func() { _ = db.Close() }()

	azClient, err := azure.CreateClient(&cfg.Azure)
	if err != nil {
		return fmt.Errorf("failed to create Azure client: %w", err)
	}

	client := azure.NewClient(azClient)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	syncer := sync.New(cfg, client, db, log)

	go func() {
		<-sigChan
		log.Info("Received interrupt signal, stopping...")
		syncer.Stop()
	}()

	for {
		if err := syncer.Start(); err != nil {
			log.Errorw("Sync failed", "error", err)
			if !cfg.Watch.Enabled {
				return err
			}
		}

		if !cfg.Watch.Enabled {
			break
		}

		log.Infow("Watch mode: sleeping", "interval", cfg.Watch.Interval)
		time.Sleep(cfg.Watch.Interval)
	}

	return nil
}
