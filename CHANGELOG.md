# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release with core synchronization functionality
- Multiple Azure authentication methods (connection string, account key, managed identity, service principal, Azure CLI)
- SQLite-based state tracking for incremental sync
- Concurrent download workers with configurable parallelism
- Watch mode for continuous synchronization
- Configuration via YAML files, environment variables, or CLI flags
- MD5 checksum verification
- Graceful shutdown on interrupts
- Comprehensive CLI with `sync`, `init`, and `status` commands
- Structured logging with configurable levels and formats
- Docker support with multi-architecture images
- Complete test suite (unit, integration, E2E)
- Extensive documentation

### Changed
- N/A (initial release)

### Deprecated
- N/A (initial release)

### Removed
- N/A (initial release)

### Fixed
- N/A (initial release)

### Security
- N/A (initial release)

---

## [1.0.0] - 2024-11-22

### Added
- Core synchronization engine with worker pool architecture
- Azure Blob Storage client with multi-auth support
- SQLite state database with WAL mode
- Configuration management with Viper
- Cobra-based CLI framework
- Structured logging with Zap
- Incremental sync using ETag and timestamps
- Blob prefix filtering
- Retry logic with exponential backoff
- Error classification and tracking
- Resume capability after interruptions
- Watch mode for continuous monitoring
- Checksum verification (MD5)
- Status reporting command
- Configuration file generator
- Docker containerization
- Multi-platform binaries (Linux, macOS, Windows on amd64/arm64)
- Comprehensive documentation (README, QUICKSTART, BUILD, DEVELOPER, TESTER, RELEASE guides)
- Full test suite with 80%+ coverage target
- CI/CD pipeline with GitHub Actions
- Automated releases for test and prod branches

### Technical Details
- Go 1.21+ support
- CGO-enabled for SQLite
- Cross-platform compatibility
- Container image size: ~20MB (Alpine-based)
- Binary size: ~12MB (with SQLite)

---

## Release Types

- **Major** (X.0.0): Breaking changes, major new features
- **Minor** (x.Y.0): New features, backwards compatible
- **Patch** (x.y.Z): Bug fixes, backwards compatible
- **Pre-release** (x.y.z-test.abc1234): Test/staging releases

---

## Upgrade Guides

### Upgrading to 1.x from 0.x (Future)

When upgrading from a hypothetical 0.x version:

1. **Backup your state database**
   ```bash
   cp .sync-state.db .sync-state.db.backup
   ```

2. **Review configuration changes**
   - Check RELEASE.md for breaking changes
   - Update configuration file if needed

3. **Test with `--dry-run`** (if implemented)
   ```bash
   getblobz sync --dry-run
   ```

4. **Perform upgrade**
   ```bash
   # Download new version
   # Replace binary
   # Verify version
   getblobz --version
   ```

---

## Links

- [Repository](https://github.com/haepapa/getblobz)
- [Releases](https://github.com/haepapa/getblobz/releases)
- [Issues](https://github.com/haepapa/getblobz/issues)
- [Discussions](https://github.com/haepapa/getblobz/discussions)

---

## Contributing

See [DEVELOPERNOTES.md](DEVELOPERNOTES.md) for development guidelines.

To add a changelog entry:
1. Add changes to the [Unreleased] section
2. Follow the categories: Added, Changed, Deprecated, Removed, Fixed, Security
3. Be concise but descriptive
4. Include issue/PR numbers when applicable

Example:
```markdown
### Added
- New feature X (#123)
- Support for Y authentication method (#124)

### Fixed
- Bug in Z component (#125)
```
