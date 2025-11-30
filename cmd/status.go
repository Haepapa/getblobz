// Package cmd provides the status command for viewing sync statistics.
package cmd

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display sync statistics and current state",
	Long: `Status shows information about sync operations including:
  - Current sync status
  - Total files synced
  - Failed downloads
  - Last sync time
  - Database statistics

Examples:
  # Show status
  getblobz status

  # Show status for specific database
  getblobz status --state-db /path/to/.sync-state.db`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().String("state-db", "./.sync-state.db", "path to state database")
}

func runStatus(cmd *cobra.Command, args []string) error {
	dbPath, _ := cmd.Flags().GetString("state-db")

	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open state database: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	var totalRuns, runningRuns, completedRuns, failedRuns int
	err = sqlDB.QueryRow(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
		FROM sync_runs
	`).Scan(&totalRuns, &runningRuns, &completedRuns, &failedRuns)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to query sync runs: %w", err)
	}

	var totalBlobs, downloadedBlobs, pendingBlobs, failedBlobs, skippedBlobs int64
	err = sqlDB.QueryRow(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'downloaded' THEN 1 ELSE 0 END) as downloaded,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
			SUM(CASE WHEN status = 'skipped' THEN 1 ELSE 0 END) as skipped
		FROM blob_state
	`).Scan(&totalBlobs, &downloadedBlobs, &pendingBlobs, &failedBlobs, &skippedBlobs)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to query blob state: %w", err)
	}

	var lastCheckTime *time.Time
	var containerName string
	err = sqlDB.QueryRow(`
		SELECT container_name, last_check_time FROM sync_checkpoint WHERE id = 1
	`).Scan(&containerName, &lastCheckTime)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to query checkpoint: %w", err)
	}

	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║           getblobz - Sync Status                         ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	if containerName != "" {
		fmt.Printf("Container:     %s\n", containerName)
		if lastCheckTime != nil {
			fmt.Printf("Last Check:    %s\n", lastCheckTime.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}

	fmt.Println("Sync Runs:")
	fmt.Printf("  Total:       %d\n", totalRuns)
	fmt.Printf("  Running:     %d\n", runningRuns)
	fmt.Printf("  Completed:   %d\n", completedRuns)
	fmt.Printf("  Failed:      %d\n", failedRuns)
	fmt.Println()

	fmt.Println("Blobs:")
	fmt.Printf("  Total:       %d\n", totalBlobs)
	fmt.Printf("  Downloaded:  %d\n", downloadedBlobs)
	fmt.Printf("  Pending:     %d\n", pendingBlobs)
	fmt.Printf("  Failed:      %d\n", failedBlobs)
	fmt.Printf("  Skipped:     %d\n", skippedBlobs)
	fmt.Println()

	if failedBlobs > 0 {
		fmt.Println("Recent Failures:")
		rows, err := sqlDB.Query(`
			SELECT blob_name, error_message, last_synced_at
			FROM blob_state 
			WHERE status = 'failed'
			ORDER BY last_synced_at DESC
			LIMIT 5
		`)
		if err == nil {
			defer func() { _ = rows.Close() }()
			for rows.Next() {
				var blobName, errorMsg string
				var lastSynced *time.Time
				if err := rows.Scan(&blobName, &errorMsg, &lastSynced); err == nil {
					timeStr := "never"
					if lastSynced != nil {
						timeStr = lastSynced.Format("2006-01-02 15:04:05")
					}
					fmt.Printf("  • %s\n    Error: %s\n    Time: %s\n", blobName, errorMsg, timeStr)
				}
			}
		}
	}

	return nil
}
