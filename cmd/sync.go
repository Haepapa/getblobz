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

	syncCmd.MarkFlagRequired("container")

	viper.BindPFlag("azure.connection_string", syncCmd.Flags().Lookup("connection-string"))
	viper.BindPFlag("azure.account_name", syncCmd.Flags().Lookup("account-name"))
	viper.BindPFlag("azure.account_key", syncCmd.Flags().Lookup("account-key"))
	viper.BindPFlag("azure.use_managed_identity", syncCmd.Flags().Lookup("use-managed-identity"))
	viper.BindPFlag("azure.tenant_id", syncCmd.Flags().Lookup("tenant-id"))
	viper.BindPFlag("azure.client_id", syncCmd.Flags().Lookup("client-id"))
	viper.BindPFlag("azure.client_secret", syncCmd.Flags().Lookup("client-secret"))
	viper.BindPFlag("azure.use_azure_cli", syncCmd.Flags().Lookup("use-azure-cli"))
	viper.BindPFlag("sync.container", syncCmd.Flags().Lookup("container"))
	viper.BindPFlag("sync.output_path", syncCmd.Flags().Lookup("output-path"))
	viper.BindPFlag("sync.prefix", syncCmd.Flags().Lookup("prefix"))
	viper.BindPFlag("sync.workers", syncCmd.Flags().Lookup("workers"))
	viper.BindPFlag("sync.batch_size", syncCmd.Flags().Lookup("batch-size"))
	viper.BindPFlag("sync.skip_existing", syncCmd.Flags().Lookup("skip-existing"))
	viper.BindPFlag("sync.verify_checksums", syncCmd.Flags().Lookup("verify-checksums"))
	viper.BindPFlag("sync.force_resync", syncCmd.Flags().Lookup("force-resync"))
	viper.BindPFlag("watch.enabled", syncCmd.Flags().Lookup("watch"))
	viper.BindPFlag("watch.interval", syncCmd.Flags().Lookup("watch-interval"))
	viper.BindPFlag("state.database", syncCmd.Flags().Lookup("state-db"))
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
	defer log.Close()

	db, err := storage.Open(cfg.State.Database)
	if err != nil {
		return fmt.Errorf("failed to open state database: %w", err)
	}
	defer db.Close()

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
