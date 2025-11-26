# Changelog

All notable changes to getblobz. Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added
- Folder organization feature for large file collections
  - Sequential, partition_key, and date-based strategies
  - Optimized for low-power devices and analytics workloads

## [1.0.0] - 2024-11-22

### Added
- Initial release
- Azure Blob Storage sync with multiple authentication methods
- SQLite state tracking for incremental sync
- Concurrent downloads with configurable worker pool
- Watch mode for continuous monitoring
- Configuration via YAML, environment variables, or CLI flags
- MD5 checksum verification
- Docker support
- Multi-platform binaries
