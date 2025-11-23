// Package storage provides SQLite database operations for state management.
package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps sql.DB with application-specific operations.
type DB struct {
	db *sql.DB
}

// Open creates or opens an SQLite database at the specified path.
// It initializes the schema if needed and configures performance settings.
func Open(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	d := &DB{db: db}
	if err := d.initialize(); err != nil {
		db.Close()
		return nil, err
	}

	return d, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// initialize creates the database schema and sets performance pragmas.
func (d *DB) initialize() error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA temp_store=MEMORY",
		"PRAGMA cache_size=-64000",
	}

	for _, pragma := range pragmas {
		if _, err := d.db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	schema := `
	CREATE TABLE IF NOT EXISTS sync_runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		started_at DATETIME NOT NULL,
		completed_at DATETIME,
		status TEXT NOT NULL,
		total_files INTEGER DEFAULT 0,
		downloaded_files INTEGER DEFAULT 0,
		failed_files INTEGER DEFAULT 0,
		total_bytes INTEGER DEFAULT 0,
		error_message TEXT
	);

	CREATE TABLE IF NOT EXISTS blob_state (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		blob_name TEXT NOT NULL UNIQUE,
		blob_path TEXT NOT NULL,
		local_path TEXT NOT NULL,
		size_bytes INTEGER NOT NULL,
		content_md5 TEXT,
		last_modified DATETIME NOT NULL,
		etag TEXT NOT NULL,
		first_seen_at DATETIME NOT NULL,
		last_synced_at DATETIME,
		sync_run_id INTEGER,
		status TEXT NOT NULL,
		error_message TEXT,
		FOREIGN KEY (sync_run_id) REFERENCES sync_runs(id)
	);

	CREATE INDEX IF NOT EXISTS idx_blob_name ON blob_state(blob_name);
	CREATE INDEX IF NOT EXISTS idx_status ON blob_state(status);
	CREATE INDEX IF NOT EXISTS idx_last_synced ON blob_state(last_synced_at);
	CREATE INDEX IF NOT EXISTS idx_etag_modified ON blob_state(etag, last_modified);

	CREATE TABLE IF NOT EXISTS sync_checkpoint (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		container_name TEXT NOT NULL,
		last_check_time DATETIME NOT NULL,
		last_continuation_token TEXT,
		total_blobs_tracked INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS performance_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sync_run_id INTEGER NOT NULL,
		timestamp DATETIME NOT NULL,
		cpu_percent REAL,
		memory_mb INTEGER,
		network_mbps REAL,
		disk_io_mbps REAL,
		active_workers INTEGER,
		download_rate_files_per_sec REAL,
		download_rate_mbps REAL,
		throttled BOOLEAN DEFAULT 0,
		FOREIGN KEY (sync_run_id) REFERENCES sync_runs(id)
	);

	CREATE INDEX IF NOT EXISTS idx_perf_sync_run ON performance_metrics(sync_run_id);
	CREATE INDEX IF NOT EXISTS idx_perf_timestamp ON performance_metrics(timestamp);

	CREATE TABLE IF NOT EXISTS error_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sync_run_id INTEGER,
		timestamp DATETIME NOT NULL,
		blob_name TEXT NOT NULL,
		error_type TEXT NOT NULL,
		error_message TEXT NOT NULL,
		retry_count INTEGER DEFAULT 0,
		resolved BOOLEAN DEFAULT 0,
		FOREIGN KEY (sync_run_id) REFERENCES sync_runs(id)
	);

	CREATE INDEX IF NOT EXISTS idx_error_timestamp ON error_log(timestamp);
	CREATE INDEX IF NOT EXISTS idx_error_resolved ON error_log(resolved);
	`

	if _, err := d.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// CreateSyncRun creates a new sync run record and returns its ID.
func (d *DB) CreateSyncRun() (int64, error) {
	result, err := d.db.Exec(
		"INSERT INTO sync_runs (started_at, status) VALUES (?, ?)",
		time.Now(), SyncStatusRunning,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create sync run: %w", err)
	}

	return result.LastInsertId()
}

// UpdateSyncRun updates an existing sync run record.
func (d *DB) UpdateSyncRun(run *SyncRun) error {
	_, err := d.db.Exec(`
		UPDATE sync_runs 
		SET completed_at = ?, status = ?, total_files = ?, 
		    downloaded_files = ?, failed_files = ?, total_bytes = ?, error_message = ?
		WHERE id = ?`,
		run.CompletedAt, run.Status, run.TotalFiles,
		run.DownloadedFiles, run.FailedFiles, run.TotalBytes, run.ErrorMessage,
		run.ID,
	)
	return err
}

// GetSyncRun retrieves a sync run by ID.
func (d *DB) GetSyncRun(id int64) (*SyncRun, error) {
	run := &SyncRun{}
	err := d.db.QueryRow(`
		SELECT id, started_at, completed_at, status, total_files, 
		       downloaded_files, failed_files, total_bytes, error_message
		FROM sync_runs WHERE id = ?`, id,
	).Scan(
		&run.ID, &run.StartedAt, &run.CompletedAt, &run.Status,
		&run.TotalFiles, &run.DownloadedFiles, &run.FailedFiles,
		&run.TotalBytes, &run.ErrorMessage,
	)
	if err != nil {
		return nil, err
	}
	return run, nil
}

// UpsertBlobState inserts or updates a blob state record.
func (d *DB) UpsertBlobState(blob *BlobState) error {
	_, err := d.db.Exec(`
		INSERT INTO blob_state 
		(blob_name, blob_path, local_path, size_bytes, content_md5, last_modified, 
		 etag, first_seen_at, last_synced_at, sync_run_id, status, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(blob_name) DO UPDATE SET
		blob_path = excluded.blob_path,
		local_path = excluded.local_path,
		size_bytes = excluded.size_bytes,
		content_md5 = excluded.content_md5,
		last_modified = excluded.last_modified,
		etag = excluded.etag,
		last_synced_at = excluded.last_synced_at,
		sync_run_id = excluded.sync_run_id,
		status = excluded.status,
		error_message = excluded.error_message`,
		blob.BlobName, blob.BlobPath, blob.LocalPath, blob.SizeBytes, blob.ContentMD5,
		blob.LastModified, blob.ETag, blob.FirstSeenAt, blob.LastSyncedAt,
		blob.SyncRunID, blob.Status, blob.ErrorMessage,
	)
	return err
}

// GetBlobState retrieves a blob state by blob name.
func (d *DB) GetBlobState(blobName string) (*BlobState, error) {
	blob := &BlobState{}
	err := d.db.QueryRow(`
		SELECT id, blob_name, blob_path, local_path, size_bytes, content_md5, 
		       last_modified, etag, first_seen_at, last_synced_at, sync_run_id, 
		       status, error_message
		FROM blob_state WHERE blob_name = ?`, blobName,
	).Scan(
		&blob.ID, &blob.BlobName, &blob.BlobPath, &blob.LocalPath, &blob.SizeBytes,
		&blob.ContentMD5, &blob.LastModified, &blob.ETag, &blob.FirstSeenAt,
		&blob.LastSyncedAt, &blob.SyncRunID, &blob.Status, &blob.ErrorMessage,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return blob, nil
}

// GetPendingBlobs returns all blobs with pending status.
func (d *DB) GetPendingBlobs() ([]*BlobState, error) {
	rows, err := d.db.Query(`
		SELECT id, blob_name, blob_path, local_path, size_bytes, content_md5, 
		       last_modified, etag, first_seen_at, last_synced_at, sync_run_id, 
		       status, error_message
		FROM blob_state WHERE status = ?`, BlobStatusPending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blobs []*BlobState
	for rows.Next() {
		blob := &BlobState{}
		if err := rows.Scan(
			&blob.ID, &blob.BlobName, &blob.BlobPath, &blob.LocalPath, &blob.SizeBytes,
			&blob.ContentMD5, &blob.LastModified, &blob.ETag, &blob.FirstSeenAt,
			&blob.LastSyncedAt, &blob.SyncRunID, &blob.Status, &blob.ErrorMessage,
		); err != nil {
			return nil, err
		}
		blobs = append(blobs, blob)
	}

	return blobs, rows.Err()
}

// RecordError logs an error to the error_log table.
func (d *DB) RecordError(syncRunID *int64, blobName, errorType, errorMessage string, retryCount int) error {
	_, err := d.db.Exec(`
		INSERT INTO error_log (sync_run_id, timestamp, blob_name, error_type, error_message, retry_count)
		VALUES (?, ?, ?, ?, ?, ?)`,
		syncRunID, time.Now(), blobName, errorType, errorMessage, retryCount,
	)
	return err
}

// RecordMetric records a performance metric snapshot.
func (d *DB) RecordMetric(metric *PerformanceMetric) error {
	_, err := d.db.Exec(`
		INSERT INTO performance_metrics 
		(sync_run_id, timestamp, cpu_percent, memory_mb, network_mbps, disk_io_mbps,
		 active_workers, download_rate_files_per_sec, download_rate_mbps, throttled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		metric.SyncRunID, metric.Timestamp, metric.CPUPercent, metric.MemoryMB,
		metric.NetworkMbps, metric.DiskIOMbps, metric.ActiveWorkers,
		metric.DownloadRateFilesPerSec, metric.DownloadRateMbps, metric.Throttled,
	)
	return err
}

// UpdateCheckpoint updates or creates the sync checkpoint.
func (d *DB) UpdateCheckpoint(containerName string, continuationToken *string) error {
	_, err := d.db.Exec(`
		INSERT INTO sync_checkpoint (id, container_name, last_check_time, last_continuation_token)
		VALUES (1, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		container_name = excluded.container_name,
		last_check_time = excluded.last_check_time,
		last_continuation_token = excluded.last_continuation_token`,
		containerName, time.Now(), continuationToken,
	)
	return err
}

// GetCheckpoint retrieves the current sync checkpoint.
func (d *DB) GetCheckpoint() (*SyncCheckpoint, error) {
	cp := &SyncCheckpoint{}
	err := d.db.QueryRow(`
		SELECT id, container_name, last_check_time, last_continuation_token, total_blobs_tracked
		FROM sync_checkpoint WHERE id = 1`,
	).Scan(&cp.ID, &cp.ContainerName, &cp.LastCheckTime, &cp.LastContinuationToken, &cp.TotalBlobsTracked)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cp, nil
}
