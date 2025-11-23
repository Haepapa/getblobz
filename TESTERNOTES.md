# getblobz - Testing Strategy and Implementation Guide

## Overview

This document provides guidance for implementing automated testing for the getblobz application. It covers unit tests, integration tests, end-to-end tests, and CI/CD integration with GitHub Actions.

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Testing Pyramid](#testing-pyramid)
3. [Test Environment Setup](#test-environment-setup)
4. [Unit Tests](#unit-tests)
5. [Integration Tests](#integration-tests)
6. [End-to-End Tests](#end-to-end-tests)
7. [Test Data Management](#test-data-management)
8. [Mocking Strategy](#mocking-strategy)
9. [Code Coverage](#code-coverage)
10. [Local Development Testing](#local-development-testing)
11. [GitHub Actions CI/CD](#github-actions-cicd)
12. [Performance Testing](#performance-testing)
13. [Test Maintenance](#test-maintenance)

---

## Testing Philosophy

### Goals

- **Reliability**: Ensure getblobz works correctly across diverse scenarios
- **Maintainability**: Tests should be easy to understand and update
- **Speed**: Fast feedback loop for developers
- **Coverage**: High code coverage without sacrificing quality
- **Confidence**: Tests should catch regressions before production

### Principles

1. **Test Pyramid**: Many unit tests, fewer integration tests, minimal e2e tests
2. **Test Isolation**: Each test should be independent
3. **Clear Naming**: Test names describe what they test
4. **Arrange-Act-Assert**: Follow AAA pattern consistently
5. **DRY with Caution**: Share setup code but keep test logic explicit

---

## Testing Pyramid

```
                    ╱╲
                   ╱  ╲
                  ╱ E2E╲          10% - Full application tests
                 ╱      ╲
                ╱────────╲
               ╱          ╲
              ╱Integration╲      30% - Component integration
             ╱              ╲
            ╱────────────────╲
           ╱                  ╲
          ╱    Unit Tests      ╲   60% - Individual functions
         ╱                      ╲
        ╱────────────────────────╲
```

### Distribution

- **Unit Tests**: ~60% of test suite
  - Fast execution (<1s per test)
  - No external dependencies
  - Test individual functions and methods

- **Integration Tests**: ~30% of test suite
  - Medium execution time (1-5s per test)
  - Test component interactions
  - Use Azurite for Azure Storage emulation

- **End-to-End Tests**: ~10% of test suite
  - Slower execution (5-30s per test)
  - Test complete workflows
  - Optional real Azure Storage tests

---

## Test Environment Setup

### Prerequisites

```bash
# Install Go 1.21+
go version

# Install testing tools
go install github.com/golang/mock/mockgen@latest
go install github.com/vektra/mockery/v2@latest
go install gotest.tools/gotestsum@latest

# Install Azurite (Azure Storage Emulator)
npm install -g azurite

# Install test coverage tools
go install github.com/axw/gocov/gocov@latest
go install github.com/AlekSi/gocov-xml@latest
```

### Directory Structure

```
getblobz/
├── cmd/
│   ├── sync_test.go
│   ├── status_test.go
│   └── init_test.go
├── internal/
│   ├── azure/
│   │   ├── auth_test.go
│   │   ├── client_test.go
│   │   └── mocks/
│   │       └── mock_client.go
│   ├── config/
│   │   └── config_test.go
│   ├── storage/
│   │   ├── db_test.go
│   │   ├── models_test.go
│   │   └── testdata/
│   │       └── test_db.sql
│   └── sync/
│       ├── syncer_test.go
│       ├── worker_test.go
│       └── mocks/
├── pkg/
│   └── logger/
│       └── logger_test.go
├── test/
│   ├── integration/
│   │   ├── sync_integration_test.go
│   │   ├── azure_integration_test.go
│   │   └── helpers.go
│   ├── e2e/
│   │   ├── full_sync_test.go
│   │   ├── watch_mode_test.go
│   │   └── fixtures/
│   └── testdata/
│       ├── config/
│       │   ├── valid.yaml
│       │   └── invalid.yaml
│       └── blobs/
│           ├── small_file.txt
│           └── large_file.bin
└── scripts/
    ├── test-local.sh
    ├── test-integration.sh
    └── start-azurite.sh
```

---

## Unit Tests

Unit tests verify individual functions and methods in isolation.

### Package: internal/config

#### config_test.go

```go
package config

import (
	"os"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	
	// Assert default values
	if cfg.Sync.Workers != 10 {
		t.Errorf("Expected workers=10, got %d", cfg.Sync.Workers)
	}
	
	if cfg.Sync.BatchSize != 5000 {
		t.Errorf("Expected batch_size=5000, got %d", cfg.Sync.BatchSize)
	}
	
	if cfg.Watch.Interval != 5*time.Minute {
		t.Errorf("Expected interval=5m, got %v", cfg.Watch.Interval)
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		Azure: AzureConfig{
			ConnectionString: "DefaultEndpointsProtocol=https;...",
		},
		Sync: SyncConfig{
			Container: "test-container",
			Workers:   10,
			BatchSize: 5000,
		},
		Performance: PerformanceConfig{
			MaxCPUPercent:     80,
			ThrottleThreshold: 0.8,
		},
	}
	
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Expected valid config, got error: %v", err)
	}
}

func TestValidate_MissingContainer(t *testing.T) {
	cfg := &Config{
		Azure: AzureConfig{
			ConnectionString: "DefaultEndpointsProtocol=https;...",
		},
		Sync: SyncConfig{
			Container: "", // Missing
		},
	}
	
	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for missing container")
	}
}

func TestValidate_InvalidWorkers(t *testing.T) {
	tests := []struct {
		name    string
		workers int
		wantErr bool
	}{
		{"Zero workers", 0, true},
		{"Valid workers", 10, false},
		{"Max workers", 100, false},
		{"Too many workers", 101, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Azure: AzureConfig{ConnectionString: "test"},
				Sync: SyncConfig{
					Container: "test",
					Workers:   tt.workers,
					BatchSize: 5000,
				},
				Performance: PerformanceConfig{
					MaxCPUPercent:     80,
					ThrottleThreshold: 0.8,
				},
			}
			
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	// Test explicit path
	path := GetConfigPath("/custom/path.yaml")
	if path != "/custom/path.yaml" {
		t.Errorf("Expected explicit path, got %s", path)
	}
	
	// Test current directory discovery
	// Create temp config file
	tmpFile := "getblobz.yaml"
	defer os.Remove(tmpFile)
	
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	
	path = GetConfigPath("")
	if path != "./getblobz.yaml" {
		t.Errorf("Expected current dir path, got %s", path)
	}
}
```

### Package: internal/storage

#### db_test.go

```go
package storage

import (
	"os"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	// Create temp database
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	
	db, err := Open(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatal(err)
	}
	
	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}
	
	return db, cleanup
}

func TestCreateSyncRun(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	runID, err := db.CreateSyncRun()
	if err != nil {
		t.Fatalf("CreateSyncRun failed: %v", err)
	}
	
	if runID == 0 {
		t.Error("Expected non-zero run ID")
	}
	
	// Verify run exists
	run, err := db.GetSyncRun(runID)
	if err != nil {
		t.Fatalf("GetSyncRun failed: %v", err)
	}
	
	if run.Status != SyncStatusRunning {
		t.Errorf("Expected status=%s, got %s", SyncStatusRunning, run.Status)
	}
}

func TestUpsertBlobState(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	now := time.Now()
	blob := &BlobState{
		BlobName:     "test-blob.txt",
		BlobPath:     "path/to/blob.txt",
		LocalPath:    "/local/path/blob.txt",
		SizeBytes:    1024,
		ETag:         "abc123",
		LastModified: now,
		FirstSeenAt:  now,
		Status:       BlobStatusPending,
	}
	
	// Insert
	err := db.UpsertBlobState(blob)
	if err != nil {
		t.Fatalf("UpsertBlobState (insert) failed: %v", err)
	}
	
	// Retrieve
	retrieved, err := db.GetBlobState("test-blob.txt")
	if err != nil {
		t.Fatalf("GetBlobState failed: %v", err)
	}
	
	if retrieved.BlobName != blob.BlobName {
		t.Errorf("Expected blob name %s, got %s", blob.BlobName, retrieved.BlobName)
	}
	
	// Update
	blob.Status = BlobStatusDownloaded
	blob.ETag = "xyz789"
	
	err = db.UpsertBlobState(blob)
	if err != nil {
		t.Fatalf("UpsertBlobState (update) failed: %v", err)
	}
	
	// Verify update
	updated, err := db.GetBlobState("test-blob.txt")
	if err != nil {
		t.Fatal(err)
	}
	
	if updated.Status != BlobStatusDownloaded {
		t.Errorf("Expected status=%s, got %s", BlobStatusDownloaded, updated.Status)
	}
	
	if updated.ETag != "xyz789" {
		t.Errorf("Expected etag=xyz789, got %s", updated.ETag)
	}
}

func TestGetPendingBlobs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	now := time.Now()
	
	// Insert test blobs
	blobs := []*BlobState{
		{
			BlobName:     "blob1.txt",
			BlobPath:     "blob1.txt",
			LocalPath:    "/local/blob1.txt",
			SizeBytes:    100,
			ETag:         "etag1",
			LastModified: now,
			FirstSeenAt:  now,
			Status:       BlobStatusPending,
		},
		{
			BlobName:     "blob2.txt",
			BlobPath:     "blob2.txt",
			LocalPath:    "/local/blob2.txt",
			SizeBytes:    200,
			ETag:         "etag2",
			LastModified: now,
			FirstSeenAt:  now,
			Status:       BlobStatusDownloaded,
		},
		{
			BlobName:     "blob3.txt",
			BlobPath:     "blob3.txt",
			LocalPath:    "/local/blob3.txt",
			SizeBytes:    300,
			ETag:         "etag3",
			LastModified: now,
			FirstSeenAt:  now,
			Status:       BlobStatusPending,
		},
	}
	
	for _, blob := range blobs {
		if err := db.UpsertBlobState(blob); err != nil {
			t.Fatal(err)
		}
	}
	
	// Get pending blobs
	pending, err := db.GetPendingBlobs()
	if err != nil {
		t.Fatalf("GetPendingBlobs failed: %v", err)
	}
	
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending blobs, got %d", len(pending))
	}
}

func TestRecordError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	runID := int64(1)
	err := db.RecordError(
		&runID,
		"test-blob.txt",
		ErrorTypeNetwork,
		"Connection timeout",
		1,
	)
	
	if err != nil {
		t.Fatalf("RecordError failed: %v", err)
	}
}
```

#### models_test.go

```go
package storage

import "testing"

func TestStatusConstants(t *testing.T) {
	// Verify status constants are defined
	statuses := []string{
		SyncStatusRunning,
		SyncStatusCompleted,
		SyncStatusFailed,
		SyncStatusInterrupted,
	}
	
	for _, status := range statuses {
		if status == "" {
			t.Error("Status constant is empty")
		}
	}
}

func TestBlobStatusConstants(t *testing.T) {
	statuses := []string{
		BlobStatusPending,
		BlobStatusDownloaded,
		BlobStatusFailed,
		BlobStatusSkipped,
	}
	
	for _, status := range statuses {
		if status == "" {
			t.Error("Blob status constant is empty")
		}
	}
}

func TestErrorTypeConstants(t *testing.T) {
	types := []string{
		ErrorTypeNetwork,
		ErrorTypeChecksum,
		ErrorTypeDisk,
		ErrorTypeAuth,
		ErrorTypeUnknown,
	}
	
	for _, errType := range types {
		if errType == "" {
			t.Error("Error type constant is empty")
		}
	}
}
```

### Package: internal/azure

#### auth_test.go

```go
package azure

import (
	"testing"
	
	"github.com/haepapa/getblobz/internal/config"
)

func TestCreateClient_ConnectionString(t *testing.T) {
	cfg := &config.AzureConfig{
		ConnectionString: "DefaultEndpointsProtocol=https;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;",
	}
	
	client, err := CreateClient(cfg)
	if err != nil {
		t.Fatalf("CreateClient failed: %v", err)
	}
	
	if client == nil {
		t.Error("Expected non-nil client")
	}
}

func TestCreateClient_NoAuth(t *testing.T) {
	cfg := &config.AzureConfig{}
	
	_, err := CreateClient(cfg)
	if err == nil {
		t.Error("Expected error for missing authentication")
	}
}

func TestCreateClient_AccountKey(t *testing.T) {
	cfg := &config.AzureConfig{
		AccountName: "testaccount",
		AccountKey:  "dGVzdGtleQ==", // base64 encoded "testkey"
	}
	
	// This will fail connection but should create client
	_, err := CreateClient(cfg)
	// We expect no error from client creation
	// Connection errors happen at usage time
	if err != nil && err.Error() != "failed to create shared key credential: decode account key: illegal base64 data at input byte 8" {
		t.Fatalf("Unexpected error: %v", err)
	}
}
```

#### client_test.go

```go
package azure

import (
	"testing"
)

func TestBlobInfo(t *testing.T) {
	info := &BlobInfo{
		Name:         "test.txt",
		Path:         "folder/test.txt",
		Size:         1024,
		ETag:         "abc123",
		LastModified: "2024-01-01T00:00:00Z",
	}
	
	if info.Name != "test.txt" {
		t.Errorf("Expected name=test.txt, got %s", info.Name)
	}
	
	if info.Size != 1024 {
		t.Errorf("Expected size=1024, got %d", info.Size)
	}
}
```

### Package: internal/sync

#### worker_test.go

```go
package sync

import (
	"testing"
	
	"github.com/haepapa/getblobz/internal/storage"
)

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Checksum error",
			err:      fmt.Errorf("checksum mismatch"),
			expected: storage.ErrorTypeChecksum,
		},
		{
			name:     "Network error",
			err:      fmt.Errorf("connection timeout"),
			expected: storage.ErrorTypeNetwork,
		},
		{
			name:     "Disk error",
			err:      fmt.Errorf("disk full"),
			expected: storage.ErrorTypeDisk,
		},
		{
			name:     "Auth error",
			err:      fmt.Errorf("unauthorized access"),
			expected: storage.ErrorTypeAuth,
		},
		{
			name:     "Unknown error",
			err:      fmt.Errorf("something went wrong"),
			expected: storage.ErrorTypeUnknown,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Network error is retryable",
			err:      fmt.Errorf("connection timeout"),
			expected: true,
		},
		{
			name:     "Checksum error is retryable",
			err:      fmt.Errorf("checksum mismatch"),
			expected: true,
		},
		{
			name:     "Disk error is not retryable",
			err:      fmt.Errorf("disk full"),
			expected: false,
		},
		{
			name:     "Auth error is not retryable",
			err:      fmt.Errorf("unauthorized"),
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryable(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"checksum error", "checksum", true},
		{"", "test", false},
	}
	
	for _, tt := range tests {
		got := contains(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}
```

### Package: pkg/logger

#### logger_test.go

```go
package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	cfg := Config{
		Level:  "info",
		Format: "text",
	}
	
	log, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	
	if log == nil {
		t.Error("Expected non-nil logger")
	}
	
	defer log.Close()
}

func TestNew_InvalidLevel(t *testing.T) {
	cfg := Config{
		Level:  "invalid",
		Format: "text",
	}
	
	log, err := New(cfg)
	if err != nil {
		t.Fatalf("Expected default level on invalid: %v", err)
	}
	
	if log == nil {
		t.Error("Expected logger with default level")
	}
	
	defer log.Close()
}

func TestNew_JSONFormat(t *testing.T) {
	cfg := Config{
		Level:  "debug",
		Format: "json",
	}
	
	log, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	
	if log == nil {
		t.Error("Expected non-nil logger")
	}
	
	defer log.Close()
}
```

---

## Integration Tests

Integration tests verify component interactions with real dependencies like Azurite.

### Setup: test/integration/helpers.go

```go
package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

const (
	AzuriteConnectionString = "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;"
)

// StartAzurite starts the Azurite emulator for testing
func StartAzurite(t *testing.T) func() {
	// Check if azurite is installed
	if _, err := exec.LookPath("azurite"); err != nil {
		t.Skip("Azurite not installed, skipping integration test")
	}
	
	// Start azurite
	cmd := exec.Command("azurite", "--silent", "--location", t.TempDir())
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start azurite: %v", err)
	}
	
	// Wait for azurite to be ready
	time.Sleep(2 * time.Second)
	
	// Return cleanup function
	return func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}
}

// CreateTestContainer creates a container in Azurite for testing
func CreateTestContainer(t *testing.T, containerName string) {
	// Implementation using Azure SDK
	// Creates container with test data
}

// UploadTestBlobs uploads test files to a container
func UploadTestBlobs(t *testing.T, containerName string, count int) {
	// Implementation uploads test files
}
```

### test/integration/azure_integration_test.go

```go
// +build integration

package integration

import (
	"context"
	"testing"
	
	"github.com/haepapa/getblobz/internal/azure"
	"github.com/haepapa/getblobz/internal/config"
)

func TestAzureClient_ListBlobs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	cleanup := StartAzurite(t)
	defer cleanup()
	
	// Create client
	cfg := &config.AzureConfig{
		ConnectionString: AzuriteConnectionString,
	}
	
	azClient, err := azure.CreateClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	client := azure.NewClient(azClient)
	
	// Create test container
	containerName := "test-container"
	CreateTestContainer(t, containerName)
	
	// Upload test blobs
	UploadTestBlobs(t, containerName, 5)
	
	// List blobs
	ctx := context.Background()
	blobs, _, err := client.ListBlobs(ctx, containerName, "", 100)
	if err != nil {
		t.Fatalf("ListBlobs failed: %v", err)
	}
	
	if len(blobs) != 5 {
		t.Errorf("Expected 5 blobs, got %d", len(blobs))
	}
}

func TestAzureClient_DownloadBlob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	cleanup := StartAzurite(t)
	defer cleanup()
	
	cfg := &config.AzureConfig{
		ConnectionString: AzuriteConnectionString,
	}
	
	azClient, err := azure.CreateClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	client := azure.NewClient(azClient)
	
	// Create test container and upload blob
	containerName := "test-container"
	blobName := "test.txt"
	CreateTestContainer(t, containerName)
	// Upload specific test blob
	
	// Download blob
	ctx := context.Background()
	tmpFile, err := os.CreateTemp("", "download-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	
	err = client.DownloadBlob(ctx, containerName, blobName, tmpFile)
	if err != nil {
		t.Fatalf("DownloadBlob failed: %v", err)
	}
	
	// Verify download
	stat, err := tmpFile.Stat()
	if err != nil {
		t.Fatal(err)
	}
	
	if stat.Size() == 0 {
		t.Error("Downloaded file is empty")
	}
}
```

### test/integration/sync_integration_test.go

```go
// +build integration

package integration

import (
	"os"
	"testing"
	"time"
	
	"github.com/haepapa/getblobz/internal/azure"
	"github.com/haepapa/getblobz/internal/config"
	"github.com/haepapa/getblobz/internal/storage"
	"github.com/haepapa/getblobz/internal/sync"
	"github.com/haepapa/getblobz/pkg/logger"
)

func TestFullSyncWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	cleanup := StartAzurite(t)
	defer cleanup()
	
	// Setup
	containerName := "test-sync"
	CreateTestContainer(t, containerName)
	UploadTestBlobs(t, containerName, 10)
	
	// Create temp output directory
	outputDir, err := os.MkdirTemp("", "sync-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outputDir)
	
	// Create temp database
	dbFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	dbFile.Close()
	defer os.Remove(dbFile.Name())
	
	// Setup configuration
	cfg := &config.Config{
		Azure: config.AzureConfig{
			ConnectionString: AzuriteConnectionString,
		},
		Sync: config.SyncConfig{
			Container:       containerName,
			OutputPath:      outputDir,
			Workers:         5,
			BatchSize:       100,
			SkipExisting:    true,
			VerifyChecksums: true,
		},
		State: config.StateConfig{
			Database: dbFile.Name(),
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
	
	// Create dependencies
	log, err := logger.New(logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	
	db, err := storage.Open(cfg.State.Database)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	
	azClient, err := azure.CreateClient(&cfg.Azure)
	if err != nil {
		t.Fatal(err)
	}
	
	client := azure.NewClient(azClient)
	
	// Create syncer
	syncer := sync.New(cfg, client, db, log)
	
	// Run sync
	err = syncer.Start()
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	
	// Verify files were downloaded
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(files) != 10 {
		t.Errorf("Expected 10 files, got %d", len(files))
	}
	
	// Verify database state
	pending, err := db.GetPendingBlobs()
	if err != nil {
		t.Fatal(err)
	}
	
	if len(pending) != 0 {
		t.Errorf("Expected 0 pending blobs, got %d", len(pending))
	}
}
```

---

## End-to-End Tests

E2E tests verify complete user workflows.

### test/e2e/full_sync_test.go

```go
// +build e2e

package e2e

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestCLI_SyncCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}
	
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "getblobz-test", "../../main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("getblobz-test")
	
	// Start Azurite
	azurite := exec.Command("azurite", "--silent")
	if err := azurite.Start(); err != nil {
		t.Fatalf("Failed to start azurite: %v", err)
	}
	defer azurite.Process.Kill()
	
	time.Sleep(2 * time.Second)
	
	// Setup test container
	// ... upload test data ...
	
	// Run sync command
	outputDir, err := os.MkdirTemp("", "e2e-sync-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outputDir)
	
	syncCmd := exec.Command("./getblobz-test", "sync",
		"--container", "test-container",
		"--connection-string", "...",
		"--output-path", outputDir,
		"--workers", "5",
	)
	
	output, err := syncCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Sync command failed: %v\nOutput: %s", err, output)
	}
	
	// Verify output
	if !strings.Contains(string(output), "Sync completed") {
		t.Errorf("Expected 'Sync completed' in output, got: %s", output)
	}
	
	// Verify files downloaded
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(files) == 0 {
		t.Error("No files were downloaded")
	}
}

func TestCLI_InitCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "e2e-init-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Build binary
	cmd := exec.Command("go", "build", "-o", "getblobz-test", "../../main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	defer os.Remove("getblobz-test")
	
	// Run init command
	configPath := tmpDir + "/test-config.yaml"
	initCmd := exec.Command("./getblobz-test", "init", "--config", configPath)
	
	output, err := initCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Init command failed: %v\nOutput: %s", err, output)
	}
	
	// Verify config file created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestCLI_StatusCommand(t *testing.T) {
	// Create test database with some data
	dbFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dbFile.Name())
	
	// ... populate database ...
	
	// Build binary
	cmd := exec.Command("go", "build", "-o", "getblobz-test", "../../main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	defer os.Remove("getblobz-test")
	
	// Run status command
	statusCmd := exec.Command("./getblobz-test", "status", "--state-db", dbFile.Name())
	
	output, err := statusCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Status command failed: %v\nOutput: %s", err, output)
	}
	
	// Verify output contains expected sections
	expectedSections := []string{
		"Sync Runs:",
		"Blobs:",
	}
	
	for _, section := range expectedSections {
		if !strings.Contains(string(output), section) {
			t.Errorf("Expected '%s' in output", section)
		}
	}
}
```

### test/e2e/watch_mode_test.go

```go
// +build e2e

package e2e

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestWatchMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}
	
	// Build binary
	cmd := exec.Command("go", "build", "-o", "getblobz-test", "../../main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	defer os.Remove("getblobz-test")
	
	// Start Azurite
	// ... setup ...
	
	// Create output directory
	outputDir, err := os.MkdirTemp("", "e2e-watch-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outputDir)
	
	// Start sync in watch mode
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	syncCmd := exec.CommandContext(ctx, "./getblobz-test", "sync",
		"--container", "test-container",
		"--connection-string", "...",
		"--output-path", outputDir,
		"--watch",
		"--watch-interval", "5s",
	)
	
	// Start command
	if err := syncCmd.Start(); err != nil {
		t.Fatalf("Failed to start watch mode: %v", err)
	}
	
	// Wait a bit for first sync
	time.Sleep(3 * time.Second)
	
	// Upload new files to container
	// ... upload files ...
	
	// Wait for next sync interval
	time.Sleep(6 * time.Second)
	
	// Verify new files were downloaded
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(files) == 0 {
		t.Error("No files downloaded in watch mode")
	}
	
	// Cancel (graceful shutdown test)
	cancel()
	
	// Wait for process to exit
	err = syncCmd.Wait()
	// Context cancellation expected
	if err != nil && err.Error() != "signal: killed" {
		t.Logf("Watch mode exit: %v", err)
	}
}
```

---

## Test Data Management

### Test Fixtures

Create reusable test data in `test/testdata/`:

```
test/testdata/
├── config/
│   ├── valid-minimal.yaml
│   ├── valid-full.yaml
│   ├── invalid-missing-container.yaml
│   └── invalid-missing-auth.yaml
├── blobs/
│   ├── small.txt (1KB)
│   ├── medium.dat (1MB)
│   └── large.bin (10MB)
└── database/
    └── sample-state.sql
```

### Test Data Generators

```go
// test/testdata/generators.go
package testdata

import (
	"crypto/rand"
	"os"
)

// GenerateRandomFile creates a file with random data
func GenerateRandomFile(path string, sizeBytes int64) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	buf := make([]byte, 4096)
	remaining := sizeBytes
	
	for remaining > 0 {
		toWrite := int64(len(buf))
		if remaining < toWrite {
			toWrite = remaining
		}
		
		if _, err := rand.Read(buf[:toWrite]); err != nil {
			return err
		}
		
		if _, err := file.Write(buf[:toWrite]); err != nil {
			return err
		}
		
		remaining -= toWrite
	}
	
	return nil
}

// CreateTestConfig returns a valid test configuration
func CreateTestConfig() *config.Config {
	return &config.Config{
		Azure: config.AzureConfig{
			ConnectionString: "...",
		},
		Sync: config.SyncConfig{
			Container:  "test-container",
			OutputPath: "/tmp/test",
			Workers:    5,
			BatchSize:  100,
		},
	}
}
```

---

## Mocking Strategy

### Generate Mocks with mockgen

```bash
# Install mockgen
go install github.com/golang/mock/mockgen@latest

# Generate mocks for interfaces
mockgen -destination=internal/azure/mocks/mock_client.go \
  -package=mocks \
  github.com/haepapa/getblobz/internal/azure Client

mockgen -destination=internal/storage/mocks/mock_db.go \
  -package=mocks \
  github.com/haepapa/getblobz/internal/storage Database
```

### Using Mocks in Tests

```go
package sync

import (
	"testing"
	
	"github.com/golang/mock/gomock"
	"github.com/haepapa/getblobz/internal/azure/mocks"
)

func TestSyncer_WithMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockClient := mocks.NewMockClient(ctrl)
	
	// Set expectations
	mockClient.EXPECT().
		ListBlobs(gomock.Any(), "test-container", "", int32(100)).
		Return([]*azure.BlobInfo{
			{Name: "test.txt", Size: 100},
		}, nil, nil)
	
	// Use mock in test
	// ...
}
```

---

## Code Coverage

### Measuring Coverage

```bash
# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# View in browser
open coverage.html
```

### Coverage Goals

| Package | Target Coverage |
|---------|----------------|
| internal/config | 90% |
| internal/storage | 85% |
| internal/azure | 75% |
| internal/sync | 80% |
| pkg/logger | 85% |
| cmd/* | 60% |

### Coverage Enforcement

Create `.codecov.yml`:

```yaml
coverage:
  status:
    project:
      default:
        target: 80%
        threshold: 2%
    patch:
      default:
        target: 75%
```

---

## Local Development Testing

### Quick Test Script

Create `scripts/test-local.sh`:

```bash
#!/bin/bash
set -e

echo "Running unit tests..."
go test -v -short ./...

echo ""
echo "Running with coverage..."
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total

echo ""
echo "Running go vet..."
go vet ./...

echo ""
echo "Running go fmt check..."
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
    echo "Files need formatting:"
    echo "$unformatted"
    exit 1
fi

echo ""
echo "✓ All local tests passed!"
```

### Integration Test Script

Create `scripts/test-integration.sh`:

```bash
#!/bin/bash
set -e

echo "Starting Azurite..."
azurite --silent --location /tmp/azurite &
AZURITE_PID=$!

# Wait for Azurite to start
sleep 3

# Ensure cleanup on exit
trap "kill $AZURITE_PID 2>/dev/null || true" EXIT

echo "Running integration tests..."
go test -v -tags=integration ./test/integration/...

echo ""
echo "✓ Integration tests passed!"
```

### E2E Test Script

Create `scripts/test-e2e.sh`:

```bash
#!/bin/bash
set -e

echo "Building application..."
go build -o getblobz main.go

echo "Starting Azurite..."
azurite --silent &
AZURITE_PID=$!

sleep 3

trap "kill $AZURITE_PID 2>/dev/null || true; rm -f getblobz" EXIT

echo "Running E2E tests..."
go test -v -tags=e2e ./test/e2e/...

echo ""
echo "✓ E2E tests passed!"
```

### Make test commands easier

Create `Makefile`:

```makefile
.PHONY: test test-unit test-integration test-e2e test-all coverage

test-unit:
	@echo "Running unit tests..."
	@go test -v -short ./...

test-integration:
	@./scripts/test-integration.sh

test-e2e:
	@./scripts/test-e2e.sh

test-all: test-unit test-integration test-e2e

coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	@golangci-lint run

fmt:
	@go fmt ./...
	@gofmt -w .

vet:
	@go vet ./...

clean:
	@rm -f coverage.out coverage.html
	@rm -f getblobz getblobz-test
```

Usage:

```bash
make test-unit           # Run unit tests only
make test-integration    # Run integration tests
make test-e2e           # Run E2E tests
make test-all           # Run all tests
make coverage           # Generate coverage report
```

---

## GitHub Actions CI/CD

### Workflow: .github/workflows/test.yml

```yaml
name: Test Suite

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  unit-tests:
    name: Unit Tests
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21, 1.22]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
      
      - name: Download dependencies
        run: go mod download
      
      - name: Run unit tests
        run: go test -v -short -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          flags: unittests
          name: codecov-unit
  
  integration-tests:
    name: Integration Tests
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.21
      
      - name: Setup Node for Azurite
        uses: actions/setup-node@v4
        with:
          node-version: '18'
      
      - name: Install Azurite
        run: npm install -g azurite
      
      - name: Start Azurite
        run: |
          azurite --silent --location /tmp/azurite &
          sleep 3
      
      - name: Run integration tests
        run: go test -v -tags=integration ./test/integration/...
  
  e2e-tests:
    name: End-to-End Tests
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.21
      
      - name: Setup Node for Azurite
        uses: actions/setup-node@v4
        with:
          node-version: '18'
      
      - name: Install Azurite
        run: npm install -g azurite
      
      - name: Build application
        run: go build -o getblobz main.go
      
      - name: Run E2E tests
        run: |
          azurite --silent &
          sleep 3
          go test -v -tags=e2e -timeout=5m ./test/e2e/...
  
  lint:
    name: Lint
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.21
      
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m
  
  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [unit-tests, integration-tests, lint]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.21
      
      - name: Build for multiple platforms
        run: |
          GOOS=linux GOARCH=amd64 go build -o getblobz-linux-amd64 main.go
          GOOS=darwin GOARCH=amd64 go build -o getblobz-darwin-amd64 main.go
          GOOS=darwin GOARCH=arm64 go build -o getblobz-darwin-arm64 main.go
          GOOS=windows GOARCH=amd64 go build -o getblobz-windows-amd64.exe main.go
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: binaries
          path: getblobz-*
```

### Workflow: .github/workflows/coverage.yml

```yaml
name: Coverage

on:
  push:
    branches: [ main ]

jobs:
  coverage:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.21
      
      - name: Run tests with coverage
        run: |
          go test -v -coverprofile=coverage.out -covermode=atomic ./...
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          flags: full
          fail_ci_if_error: true
      
      - name: Check coverage threshold
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          threshold=75
          if (( $(echo "$coverage < $threshold" | bc -l) )); then
            echo "Coverage $coverage% is below threshold $threshold%"
            exit 1
          fi
```

### Workflow: .github/workflows/release.yml

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: 1.21
      
      - name: Run all tests
        run: |
          go test -v ./...
  
  build-and-release:
    needs: test
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.21
      
      - name: Build binaries
        run: |
          mkdir -p dist
          GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/getblobz-linux-amd64 main.go
          GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/getblobz-linux-arm64 main.go
          GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/getblobz-darwin-amd64 main.go
          GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/getblobz-darwin-arm64 main.go
          GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/getblobz-windows-amd64.exe main.go
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## Performance Testing

### Benchmark Tests

Create `internal/sync/syncer_bench_test.go`:

```go
package sync

import (
	"testing"
)

func BenchmarkWorkerProcessing(b *testing.B) {
	// Setup
	// ...
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Process blob
	}
}

func BenchmarkDatabaseInsert(b *testing.B) {
	db, cleanup := setupTestDB(b)
	defer cleanup()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blob := &storage.BlobState{
			BlobName:  fmt.Sprintf("blob-%d.txt", i),
			// ... other fields
		}
		db.UpsertBlobState(blob)
	}
}
```

### Load Testing

Create `test/load/load_test.go`:

```go
// +build load

package load

import (
	"sync"
	"testing"
)

func TestConcurrentDownloads(t *testing.T) {
	concurrency := 100
	blobsPerWorker := 10
	
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			// Simulate downloading blobs
			for j := 0; j < blobsPerWorker; j++ {
				// Download logic
			}
		}(i)
	}
	
	wg.Wait()
}
```

---

## Test Maintenance

### Best Practices

1. **Keep Tests Fast**
   - Unit tests should run in < 1s
   - Use `-short` flag for quick feedback
   - Mock external dependencies

2. **Avoid Flaky Tests**
   - Don't rely on timing
   - Use proper synchronization
   - Clean up resources

3. **Test Naming Convention**
   ```go
   func TestFunctionName_Scenario_ExpectedBehavior(t *testing.T)
   // Example:
   func TestValidate_MissingContainer_ReturnsError(t *testing.T)
   ```

4. **Table-Driven Tests**
   ```go
   tests := []struct {
       name    string
       input   interface{}
       want    interface{}
       wantErr bool
   }{
       // test cases
   }
   ```

5. **Cleanup Resources**
   ```go
   defer cleanup()
   t.Cleanup(func() {
       // cleanup code
   })
   ```

### Test Code Review Checklist

- [ ] Tests are independent and isolated
- [ ] Clear test names describing what is tested
- [ ] Both happy path and error cases covered
- [ ] No hardcoded values (use constants/fixtures)
- [ ] Proper cleanup of resources
- [ ] Mocks are used appropriately
- [ ] Tests are deterministic (not flaky)
- [ ] Coverage meets threshold
- [ ] Performance tests for critical paths

---

## Summary

This testing strategy provides:

1. **Comprehensive Coverage**: Unit, integration, and E2E tests
2. **Fast Feedback**: Quick unit tests for development
3. **CI/CD Integration**: Automated testing on GitHub Actions
4. **Quality Assurance**: Coverage tracking and enforcement
5. **Maintainability**: Clear structure and best practices

### Quick Reference

```bash
# Local Development
make test-unit              # Fast unit tests
make test-integration       # Integration tests with Azurite
make test-e2e              # Full E2E tests
make coverage              # Generate coverage report

# CI/CD
# Automated on push/PR via GitHub Actions

# Coverage Goals
# Total: 80%+
# Critical packages: 85%+
```

### Next Steps for Test Developer

1. Start with unit tests for `internal/config` and `internal/storage`
2. Set up Azurite and create integration test helpers
3. Write integration tests for Azure client operations
4. Implement E2E tests for main workflows
5. Configure GitHub Actions workflows
6. Set up coverage tracking with Codecov
7. Add performance benchmarks for critical paths

---

**Document Version**: 1.0  
**Last Updated**: 2024-11-22  
**Maintained By**: Haepapa
