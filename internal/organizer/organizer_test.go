package organizer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/haepapa/getblobz/internal/config"
)

func TestOrganizer_Disabled(t *testing.T) {
	cfg := &config.FolderOrganizationConfig{
		Enabled: false,
	}

	org := New(cfg, "/data")
	path := org.GetTargetPath("blob1.txt", "files/blob1.txt")

	expected := filepath.Join("/data", "files/blob1.txt")
	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestOrganizer_Sequential(t *testing.T) {
	cfg := &config.FolderOrganizationConfig{
		Enabled:           true,
		MaxFilesPerFolder: 3,
		Strategy:          "sequential",
	}

	org := New(cfg, "/data")

	paths := []string{}
	for i := 0; i < 10; i++ {
		path := org.GetTargetPath("blob.txt", "file.txt")
		paths = append(paths, path)
	}

	if !contains(paths[0], "folder_0000") {
		t.Errorf("First file should be in folder_0000")
	}

	if !contains(paths[3], "folder_0001") {
		t.Errorf("Fourth file should be in folder_0001 due to max 3 files per folder")
	}

	if !contains(paths[9], "folder_0003") {
		t.Errorf("Tenth file should be in folder_0003")
	}
}

func TestOrganizer_PartitionKey(t *testing.T) {
	cfg := &config.FolderOrganizationConfig{
		Enabled:        true,
		Strategy:       "partition_key",
		PartitionDepth: 2,
	}

	org := New(cfg, "/data")

	path1 := org.GetTargetPath("blob1.txt", "file.txt")
	path2 := org.GetTargetPath("blob1.txt", "file.txt")

	if path1 != path2 {
		t.Errorf("Same blob name should produce same path")
	}

	path3 := org.GetTargetPath("blob2.txt", "file.txt")
	if path1 == path3 {
		t.Logf("Different blob names might produce different paths (not guaranteed)")
	}
}

func TestOrganizer_DateStrategy(t *testing.T) {
	cfg := &config.FolderOrganizationConfig{
		Enabled:  true,
		Strategy: "date",
	}

	org := New(cfg, "/data")
	path := org.GetTargetPath("blob.txt", "file.txt")

	if !contains(path, "/data/") {
		t.Errorf("Path should contain base path")
	}
}

func TestOrganizer_LoadState(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "folder_0000"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "folder_0001"), 0755)

	for i := 0; i < 5; i++ {
		os.WriteFile(filepath.Join(tmpDir, "folder_0000", "file.txt"), []byte("test"), 0644)
	}

	cfg := &config.FolderOrganizationConfig{
		Enabled:           true,
		MaxFilesPerFolder: 10,
		Strategy:          "sequential",
	}

	org := New(cfg, tmpDir)
	err := org.LoadState()
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	stats := org.GetStats()
	if stats["total_folders"].(int) < 1 {
		t.Errorf("Should have detected at least 1 folder")
	}
}

func TestOrganizer_GetStats(t *testing.T) {
	cfg := &config.FolderOrganizationConfig{
		Enabled:           true,
		MaxFilesPerFolder: 5,
		Strategy:          "sequential",
	}

	org := New(cfg, "/data")

	for i := 0; i < 7; i++ {
		org.GetTargetPath("blob.txt", "file.txt")
	}

	stats := org.GetStats()

	if !stats["enabled"].(bool) {
		t.Errorf("Organizer should be enabled")
	}

	if stats["strategy"].(string) != "sequential" {
		t.Errorf("Strategy should be sequential")
	}

	if stats["total_files"].(int) != 7 {
		t.Errorf("Expected 7 files, got %d", stats["total_files"])
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
