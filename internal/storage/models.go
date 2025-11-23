// Package storage provides database models and operations for state management.
package storage

import (
	"time"
)

// SyncRun represents a synchronisation run record.
type SyncRun struct {
	ID              int64
	StartedAt       time.Time
	CompletedAt     *time.Time
	Status          string
	TotalFiles      int64
	DownloadedFiles int64
	FailedFiles     int64
	TotalBytes      int64
	ErrorMessage    *string
}

// BlobState tracks the state of an individual blob.
type BlobState struct {
	ID           int64
	BlobName     string
	BlobPath     string
	LocalPath    string
	SizeBytes    int64
	ContentMD5   *string
	LastModified time.Time
	ETag         string
	FirstSeenAt  time.Time
	LastSyncedAt *time.Time
	SyncRunID    *int64
	Status       string
	ErrorMessage *string
}

// SyncCheckpoint stores the last known state for incremental syncing.
type SyncCheckpoint struct {
	ID                    int64
	ContainerName         string
	LastCheckTime         time.Time
	LastContinuationToken *string
	TotalBlobsTracked     int64
}

// PerformanceMetric records system performance data during sync operations.
type PerformanceMetric struct {
	ID                      int64
	SyncRunID               int64
	Timestamp               time.Time
	CPUPercent              *float64
	MemoryMB                *int64
	NetworkMbps             *float64
	DiskIOMbps              *float64
	ActiveWorkers           *int
	DownloadRateFilesPerSec *float64
	DownloadRateMbps        *float64
	Throttled               bool
}

// ErrorLog stores detailed error information for debugging.
type ErrorLog struct {
	ID           int64
	SyncRunID    *int64
	Timestamp    time.Time
	BlobName     string
	ErrorType    string
	ErrorMessage string
	RetryCount   int
	Resolved     bool
}

const (
	// SyncStatusRunning indicates an active sync operation.
	SyncStatusRunning = "running"
	// SyncStatusCompleted indicates a successfully completed sync.
	SyncStatusCompleted = "completed"
	// SyncStatusFailed indicates a failed sync operation.
	SyncStatusFailed = "failed"
	// SyncStatusInterrupted indicates an interrupted sync operation.
	SyncStatusInterrupted = "interrupted"
)

const (
	// BlobStatusPending indicates a blob waiting to be downloaded.
	BlobStatusPending = "pending"
	// BlobStatusDownloaded indicates a successfully downloaded blob.
	BlobStatusDownloaded = "downloaded"
	// BlobStatusFailed indicates a failed download attempt.
	BlobStatusFailed = "failed"
	// BlobStatusSkipped indicates a skipped blob (already exists).
	BlobStatusSkipped = "skipped"
)

const (
	// ErrorTypeNetwork indicates a network-related error.
	ErrorTypeNetwork = "network"
	// ErrorTypeChecksum indicates a checksum validation error.
	ErrorTypeChecksum = "checksum"
	// ErrorTypeDisk indicates a disk I/O error.
	ErrorTypeDisk = "disk"
	// ErrorTypeAuth indicates an authentication error.
	ErrorTypeAuth = "auth"
	// ErrorTypeUnknown indicates an unclassified error.
	ErrorTypeUnknown = "unknown"
)
