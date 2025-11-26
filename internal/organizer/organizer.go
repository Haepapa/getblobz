// Package organizer provides folder organization logic for downloaded files.
package organizer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/haepapa/getblobz/internal/config"
)

// Organizer manages folder organization for downloaded files.
type Organizer struct {
	cfg           *config.FolderOrganizationConfig
	basePath      string
	mu            sync.RWMutex
	folderCounts  map[string]int
	currentFolder string
	folderIndex   int
}

// New creates a new Organizer instance.
func New(cfg *config.FolderOrganizationConfig, basePath string) *Organizer {
	return &Organizer{
		cfg:          cfg,
		basePath:     basePath,
		folderCounts: make(map[string]int),
		folderIndex:  0,
	}
}

// GetTargetPath returns the appropriate folder path for a file based on the organization strategy.
// This method is thread-safe and ensures files are distributed according to the configured strategy.
func (o *Organizer) GetTargetPath(blobName string, blobPath string) string {
	if !o.cfg.Enabled {
		return filepath.Join(o.basePath, blobPath)
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	var folder string

	switch o.cfg.Strategy {
	case "partition_key":
		folder = o.getPartitionKeyFolder(blobName)
	case "date":
		folder = o.getDateFolder()
	case "sequential":
		folder = o.getSequentialFolder()
	default:
		folder = o.getSequentialFolder()
	}

	targetPath := filepath.Join(o.basePath, folder, blobPath)
	o.trackFile(folder)

	return targetPath
}

// getPartitionKeyFolder generates a folder path based on hash partitioning of the blob name.
// This distributes files evenly across folders using hash-based partitioning,
// which is optimal for analytics workloads like Apache Spark.
func (o *Organizer) getPartitionKeyFolder(blobName string) string {
	hash := sha256.Sum256([]byte(blobName))
	hashStr := hex.EncodeToString(hash[:])

	parts := make([]string, o.cfg.PartitionDepth)
	charsPerPart := 2

	for i := 0; i < o.cfg.PartitionDepth; i++ {
		start := i * charsPerPart
		end := start + charsPerPart
		if end > len(hashStr) {
			end = len(hashStr)
		}
		parts[i] = hashStr[start:end]
	}

	return filepath.Join(parts...)
}

// getDateFolder generates a folder path based on the current date.
// Format: YYYY/MM/DD for hierarchical date-based organization.
func (o *Organizer) getDateFolder() string {
	now := time.Now()
	return filepath.Join(
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
	)
}

// getSequentialFolder generates a sequential folder path (folder_0000, folder_0001, etc.).
// When the current folder reaches the max file limit, it automatically creates the next folder.
func (o *Organizer) getSequentialFolder() string {
	if o.currentFolder == "" || o.folderCounts[o.currentFolder] >= o.cfg.MaxFilesPerFolder {
		o.currentFolder = fmt.Sprintf("folder_%04d", o.folderIndex)
		o.folderIndex++
	}

	return o.currentFolder
}

// trackFile increments the file count for a given folder.
func (o *Organizer) trackFile(folder string) {
	o.folderCounts[folder]++
}

// LoadState loads the current state of folder organization from the filesystem.
// This scans existing folders to determine current file counts and folder indices.
func (o *Organizer) LoadState() error {
	if !o.cfg.Enabled {
		return nil
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	switch o.cfg.Strategy {
	case "sequential":
		return o.loadSequentialState()
	case "partition_key", "date":
		return o.loadPartitionedState()
	}

	return nil
}

// loadSequentialState scans for existing sequential folders and determines the next folder index.
func (o *Organizer) loadSequentialState() error {
	if _, err := os.Stat(o.basePath); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(o.basePath)
	if err != nil {
		return fmt.Errorf("failed to read base path: %w", err)
	}

	maxIndex := -1
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		var idx int
		if n, _ := fmt.Sscanf(entry.Name(), "folder_%d", &idx); n == 1 {
			if idx > maxIndex {
				maxIndex = idx
			}

			folderPath := filepath.Join(o.basePath, entry.Name())
			count, _ := countFilesInFolder(folderPath)
			o.folderCounts[entry.Name()] = count

			if count < o.cfg.MaxFilesPerFolder {
				o.currentFolder = entry.Name()
				o.folderIndex = idx
			}
		}
	}

	if maxIndex >= 0 && o.currentFolder == "" {
		o.folderIndex = maxIndex + 1
	}

	return nil
}

// loadPartitionedState scans partition-based folders and counts files per partition.
func (o *Organizer) loadPartitionedState() error {
	if _, err := os.Stat(o.basePath); os.IsNotExist(err) {
		return nil
	}

	return filepath.WalkDir(o.basePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(o.basePath, path)
		if err != nil {
			return nil
		}

		if relPath == "." {
			return nil
		}

		count, _ := countFilesInFolder(path)
		o.folderCounts[relPath] = count

		return nil
	})
}

// countFilesInFolder counts the number of files (not directories) in a folder.
func countFilesInFolder(folderPath string) (int, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	return count, nil
}

// GetStats returns statistics about the current folder organization.
func (o *Organizer) GetStats() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	stats := map[string]interface{}{
		"enabled":       o.cfg.Enabled,
		"strategy":      o.cfg.Strategy,
		"total_folders": len(o.folderCounts),
		"total_files":   0,
	}

	totalFiles := 0
	for _, count := range o.folderCounts {
		totalFiles += count
	}
	stats["total_files"] = totalFiles

	if o.cfg.Strategy == "sequential" {
		stats["current_folder"] = o.currentFolder
		stats["next_folder_index"] = o.folderIndex
	}

	return stats
}
