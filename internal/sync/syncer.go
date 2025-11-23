// Package sync implements the core synchronisation logic for getblobz.
package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/haepapa/getblobz/internal/azure"
	"github.com/haepapa/getblobz/internal/config"
	"github.com/haepapa/getblobz/internal/storage"
	"github.com/haepapa/getblobz/pkg/logger"
)

// Syncer manages the blob synchronisation process.
type Syncer struct {
	cfg    *config.Config
	client *azure.Client
	db     *storage.DB
	logger *logger.Logger

	runID   int64
	workers int
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

// New creates a new Syncer instance.
func New(cfg *config.Config, client *azure.Client, db *storage.DB, log *logger.Logger) *Syncer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Syncer{
		cfg:     cfg,
		client:  client,
		db:      db,
		logger:  log,
		workers: cfg.Sync.Workers,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start begins the synchronisation process.
// It orchestrates discovery, download, and completion phases.
func (s *Syncer) Start() error {
	var err error
	s.runID, err = s.db.CreateSyncRun()
	if err != nil {
		return fmt.Errorf("failed to create sync run: %w", err)
	}

	s.logger.Infow("Sync started",
		"container", s.cfg.Sync.Container,
		"output_path", s.cfg.Sync.OutputPath,
		"workers", s.workers,
		"run_id", s.runID,
	)

	if err := s.discovery(); err != nil {
		s.markRunFailed(err)
		return fmt.Errorf("discovery failed: %w", err)
	}

	if err := s.download(); err != nil {
		s.markRunFailed(err)
		return fmt.Errorf("download failed: %w", err)
	}

	if err := s.complete(); err != nil {
		s.markRunFailed(err)
		return fmt.Errorf("completion failed: %w", err)
	}

	return nil
}

// Stop gracefully stops the synchronisation process.
func (s *Syncer) Stop() {
	s.logger.Info("Stopping sync...")
	s.cancel()
	s.wg.Wait()
}

// discovery lists all blobs and determines which need to be downloaded.
func (s *Syncer) discovery() error {
	s.logger.Infow("Starting discovery phase", "prefix", s.cfg.Sync.Prefix)

	var totalFound int64
	var totalNew int64
	var totalChanged int64
	var totalSkipped int64

	var continuationToken *string
	batchSize := int32(s.cfg.Sync.BatchSize)

	for {
		blobs, token, err := s.client.ListBlobs(
			s.ctx,
			s.cfg.Sync.Container,
			s.cfg.Sync.Prefix,
			batchSize,
		)
		if err != nil {
			return fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, blob := range blobs {
			totalFound++

			existing, err := s.db.GetBlobState(blob.Name)
			if err != nil {
				s.logger.Warnw("Failed to get blob state", "blob", blob.Name, "error", err)
				continue
			}

			status := storage.BlobStatusPending
			isNew := existing == nil

			if !isNew {
				if !s.cfg.Sync.ForceResync {
					if existing.ETag == blob.ETag && existing.LastModified.Format("2006-01-02T15:04:05Z") == blob.LastModified {
						if s.cfg.Sync.SkipExisting {
							status = storage.BlobStatusSkipped
							totalSkipped++
						} else {
							totalChanged++
						}
					} else {
						totalChanged++
					}
				}
			} else {
				totalNew++
			}

			lastModified, _ := time.Parse("2006-01-02T15:04:05Z", blob.LastModified)
			blobState := &storage.BlobState{
				BlobName:     blob.Name,
				BlobPath:     blob.Path,
				LocalPath:    fmt.Sprintf("%s/%s", s.cfg.Sync.OutputPath, blob.Path),
				SizeBytes:    blob.Size,
				ETag:         blob.ETag,
				LastModified: lastModified,
				FirstSeenAt:  time.Now(),
				Status:       status,
			}

			if len(blob.ContentMD5) > 0 {
				md5Str := fmt.Sprintf("%x", blob.ContentMD5)
				blobState.ContentMD5 = &md5Str
			}

			if err := s.db.UpsertBlobState(blobState); err != nil {
				s.logger.Warnw("Failed to upsert blob state", "blob", blob.Name, "error", err)
			}
		}

		continuationToken = token
		if continuationToken == nil {
			break
		}

		s.logger.Infow("Discovery progress", "found", totalFound)
	}

	s.logger.Infow("Discovery completed",
		"total", totalFound,
		"new", totalNew,
		"changed", totalChanged,
		"skipped", totalSkipped,
	)

	if err := s.db.UpdateCheckpoint(s.cfg.Sync.Container, continuationToken); err != nil {
		s.logger.Warnw("Failed to update checkpoint", "error", err)
	}

	return nil
}

// download processes pending blobs using a worker pool.
func (s *Syncer) download() error {
	s.logger.Info("Starting download phase")

	pending, err := s.db.GetPendingBlobs()
	if err != nil {
		return fmt.Errorf("failed to get pending blobs: %w", err)
	}

	if len(pending) == 0 {
		s.logger.Info("No blobs to download")
		return nil
	}

	s.logger.Infow("Downloading blobs", "count", len(pending))

	blobQueue := make(chan *storage.BlobState, len(pending))
	for _, blob := range pending {
		blobQueue <- blob
	}
	close(blobQueue)

	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i, blobQueue)
	}

	s.wg.Wait()
	s.logger.Info("Download phase completed")

	return nil
}

// complete finalizes the sync run and logs statistics.
func (s *Syncer) complete() error {
	s.logger.Info("Completing sync run")

	run, err := s.db.GetSyncRun(s.runID)
	if err != nil {
		return fmt.Errorf("failed to get sync run: %w", err)
	}

	now := time.Now()
	run.CompletedAt = &now
	run.Status = storage.SyncStatusCompleted

	if err := s.db.UpdateSyncRun(run); err != nil {
		return fmt.Errorf("failed to update sync run: %w", err)
	}

	duration := run.CompletedAt.Sub(run.StartedAt)
	s.logger.Infow("Sync completed",
		"duration", duration.String(),
		"downloaded", run.DownloadedFiles,
		"failed", run.FailedFiles,
		"total_bytes", run.TotalBytes,
	)

	return nil
}

// markRunFailed marks the sync run as failed with an error message.
func (s *Syncer) markRunFailed(err error) {
	run, dbErr := s.db.GetSyncRun(s.runID)
	if dbErr != nil {
		s.logger.Errorw("Failed to get sync run for failure marking", "error", dbErr)
		return
	}

	now := time.Now()
	run.CompletedAt = &now
	run.Status = storage.SyncStatusFailed
	errMsg := err.Error()
	run.ErrorMessage = &errMsg

	if updateErr := s.db.UpdateSyncRun(run); updateErr != nil {
		s.logger.Errorw("Failed to update failed sync run", "error", updateErr)
	}
}
