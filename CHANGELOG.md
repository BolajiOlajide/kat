# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- SQLite database driver support with pure Go implementation via `modernc.org/sqlite` - no CGO required (#27)
- New `driver` config field in `kat.conf.yaml` (defaults to `postgres` for backward compatibility)
- Database connection hardening with configurable timeouts and pool management (#29)
- Configurable connect/statement timeouts and pool limits via `kat.conf.yaml` or library options (`WithDBConfig`, `WithConnectTimeout`, `WithPoolLimits`)
- Retry with exponential backoff for `Ping` on transient Postgres errors

### Changed
- Use string for MigrationLog Duration instead of `pgtype.Interval` for better cross-driver compatibility (#34)
- Bumped Go toolchain to 1.25.7 and CI action versions

### Fixed
- Preserve original Postgres URL in `ConnString()` instead of reconstructing it, preventing loss of query parameters and special characters (#35)
- Consistent table name quoting and validation across all SQL templates (#36)
- Enforce `MaxOpenConns=1` for SQLite in `NewWithDB` to prevent "database is locked" errors (#37)
- Restore `MigrationTime` as `time.Time` with smart scanning for both Postgres and SQLite (#38)
- Error on invalid driver in `SetDefault` instead of silently defaulting to Postgres (#39)
- Transaction handling bug in `WithTransact()` discovered during SQLite integration testing

## [0.1.2] - 2026-02-23

### Fixed
- Fix warning style color to yellow instead of red (#33)

## [0.1.1] - 2026-02-20

### Added
- YAML schemas and editor support for `kat.conf.yaml` and migration metadata files (#31)

### Fixed
- Fix permission denied error in `kat update` on macOS (#32)

## [0.1.0] - 2026-02-19

### Added
- Non-transactional migration support via `no_transaction` field in migration metadata (#30)
- Documentation for non-transactional migrations

## [0.0.11] - 2025-09-06

### Changed
- Updated installation command (#24)
- Updated documentation (#25)

### Fixed
- Fix critical transaction bug in `WithTransact` function (#28)

## [0.0.10] - 2025-07-19

### Added
- More API improvements with enhanced functionality

## [0.0.9] - 2025-07-19

### Added
- Go documentation for kat library
- Custom logger support for better integration

## [0.0.8] - 2025-05-29

### Fixed
- Create migration directory automatically if it doesn't exist when running `kat add`

## [0.0.7] - 2025-05-20

### Fixed
- Fix update command functionality

## [0.0.6] - 2025-05-20

### Added
- Graph-based migration system implementation
- New command for exporting migration graph
- Optimized compute definitions

### Changed
- Refactored migration metadata structure
- Improved execution flow with enhanced runner logic

### Fixed
- Fixed execution counting logic
- Fixed documentation links
- Various refactoring improvements to runner logic

## [0.0.5] - 2025-05-09

### Added
- Published kat as a Go module for programmatic usage

## [0.0.4] - 2025-04-29

### Added
- Added `update` command to automatically check for and install new versions
- Added comprehensive documentation for updating Kat and contributing to the project
- Added test coverage for update functionality

## [0.0.3] - 2025-04-21

### Changed
- Removed binary uploads from release process
- Polished documentation for better clarity

### Fixed
- Fixed architecture detection in install script to properly support different CPU architectures

## [0.0.2] - 2025-04-20

### Added
- Added description field for migration metadata
- Added command-line flags for verbosity control

### Security
- Added GPG signing for releases

## [0.0.1] - 2025-04-10

### Added
- Initial alpha release
- Basic migration functionality
- Support for PostgreSQL databases
- Simple CLI interface

[Unreleased]: https://github.com/BolajiOlajide/kat/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/BolajiOlajide/kat/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/BolajiOlajide/kat/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/BolajiOlajide/kat/compare/v0.0.11...v0.1.0
[0.0.11]: https://github.com/BolajiOlajide/kat/compare/v0.0.10...v0.0.11
[0.0.10]: https://github.com/BolajiOlajide/kat/compare/v0.0.9...v0.0.10
[0.0.9]: https://github.com/BolajiOlajide/kat/compare/v0.0.8...v0.0.9
[0.0.8]: https://github.com/BolajiOlajide/kat/compare/v0.0.7...v0.0.8
[0.0.7]: https://github.com/BolajiOlajide/kat/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/BolajiOlajide/kat/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/BolajiOlajide/kat/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/BolajiOlajide/kat/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/BolajiOlajide/kat/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/BolajiOlajide/kat/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/BolajiOlajide/kat/releases/tag/v0.0.1