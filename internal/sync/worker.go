// Package sync provides worker goroutines for concurrent blob downloads.
package sync

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/haepapa/getblobz/internal/storage"
)

const (
	maxRetries = 3
	baseDelay  = 1 * time.Second
)

// worker is a goroutine that processes blobs from the queue.
func (s *Syncer) worker(id int, queue <-chan *storage.BlobState) {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case blob, ok := <-queue:
			if !ok {
				return
			}
			s.processBlob(id, blob)
		}
	}
}

// fsUsagePercent calculates filesystem usage percent for the directory containing the target path.
func fsUsagePercent(dir string) (int, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return 0, err
	}
	// Use Bavail for non-root available blocks.
	total := float64(stat.Blocks) * float64(stat.Bsize)
	avail := float64(stat.Bavail) * float64(stat.Bsize)
	if total <= 0 {
		return 0, fmt.Errorf("invalid filesystem size")
	}
	usedPercent := int(((total - avail) / total) * 100.0)
	if usedPercent < 0 {
		usedPercent = 0
	}
	if usedPercent > 100 {
		usedPercent = 100
	}
	return usedPercent, nil
}

// processBlob downloads and saves a single blob with retry logic.
func (s *Syncer) processBlob(workerID int, blob *storage.BlobState) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			s.logger.Infow("Retrying blob download",
				"worker", workerID,
				"blob", blob.BlobName,
				"attempt", attempt+1,
				"delay", delay,
			)
			time.Sleep(delay)
		}

		// Check disk usage before attempting download
		usage, duErr := fsUsagePercent(filepath.Dir(s.cfg.Sync.OutputPath))
		if duErr == nil {
			if usage >= s.cfg.Sync.DiskStopPercent {
				s.logger.Errorw("Filesystem usage exceeded stop threshold; stopping downloads",
					"usage_percent", usage,
					"stop_percent", s.cfg.Sync.DiskStopPercent,
				)
				lastErr = fmt.Errorf("disk usage %d%% >= stop threshold %d%%", usage, s.cfg.Sync.DiskStopPercent)
				break
			}
			if usage >= s.cfg.Sync.DiskWarnPercent {
				s.logger.Warnw("Filesystem usage exceeded warn threshold",
					"usage_percent", usage,
					"warn_percent", s.cfg.Sync.DiskWarnPercent,
				)
			}
		} else {
			s.logger.Warnw("Failed to check filesystem usage", "error", duErr)
		}

		err := s.downloadBlob(workerID, blob)
		if err == nil {
			blob.Status = storage.BlobStatusDownloaded
			now := time.Now()
			blob.LastSyncedAt = &now
			blob.SyncRunID = &s.runID

			if err := s.db.UpsertBlobState(blob); err != nil {
				s.logger.Warnw("Failed to update blob state",
					"worker", workerID,
					"blob", blob.BlobName,
					"error", err,
				)
			}

			s.logger.Infow("Downloaded blob",
				"worker", workerID,
				"blob", blob.BlobName,
				"size", blob.SizeBytes,
			)
			return
		}

		lastErr = err
		errorType := classifyError(err)
		if err := s.db.RecordError(&s.runID, blob.BlobName, errorType, err.Error(), attempt); err != nil {
			s.logger.Warnw("Failed to record error", "error", err)
		}

		if !isRetryable(err) {
			break
		}
	}

	blob.Status = storage.BlobStatusFailed
	errMsg := lastErr.Error()
	blob.ErrorMessage = &errMsg

	if err := s.db.UpsertBlobState(blob); err != nil {
		s.logger.Warnw("Failed to update failed blob state",
			"worker", workerID,
			"blob", blob.BlobName,
			"error", err,
		)
	}

	s.logger.Errorw("Failed to download blob",
		"worker", workerID,
		"blob", blob.BlobName,
		"error", lastErr,
	)
}

// downloadBlob performs the actual blob download.
func (s *Syncer) downloadBlob(workerID int, blob *storage.BlobState) error {
	dir := filepath.Dir(blob.LocalPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tmpPath := blob.LocalPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var writer io.Writer = file
	var hash io.Writer

	if s.cfg.Sync.VerifyChecksums && blob.ContentMD5 != nil {
		hasher := md5.New()
		writer = io.MultiWriter(file, hasher)
		hash = hasher
	}

	err = s.client.DownloadBlob(s.ctx, s.cfg.Sync.Container, blob.BlobName, writer)
	if err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("download failed: %w", err)
	}

	if s.cfg.Sync.VerifyChecksums && blob.ContentMD5 != nil && hash != nil {
		computed := hex.EncodeToString(hash.(interface{ Sum([]byte) []byte }).Sum(nil))
		if computed != *blob.ContentMD5 {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("checksum mismatch: expected %s, got %s", *blob.ContentMD5, computed)
		}
	}

	_ = file.Close()

	if err := os.Rename(tmpPath, blob.LocalPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// classifyError categorizes errors for logging and reporting.
func classifyError(err error) string {
	if err == nil {
		return storage.ErrorTypeUnknown
	}

	errStr := err.Error()
	if contains(errStr, "checksum") || contains(errStr, "md5") {
		return storage.ErrorTypeChecksum
	}
	if contains(errStr, "network") || contains(errStr, "timeout") || contains(errStr, "connection") {
		return storage.ErrorTypeNetwork
	}
	if contains(errStr, "disk") || contains(errStr, "space") || contains(errStr, "permission") {
		return storage.ErrorTypeDisk
	}
	if contains(errStr, "auth") || contains(errStr, "unauthorized") {
		return storage.ErrorTypeAuth
	}

	return storage.ErrorTypeUnknown
}

// isRetryable determines if an error should trigger a retry.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	errType := classifyError(err)
	return errType == storage.ErrorTypeNetwork || errType == storage.ErrorTypeChecksum
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0))
}

// indexOf returns the index of substr in s, or -1 if not found.
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
