# Kat

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/BolajiOlajide/kat/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/BolajiOlajide/kat)](https://goreportcard.com/report/github.com/BolajiOlajide/kat)

Kat is a lightweight, powerful CLI tool for PostgreSQL database migrations. It allows you to manage your database schema using SQL files with a simple, intuitive workflow.

![Kat Banner](doc/assets/images/layout/logo.png)

## Features

- **Simple SQL Migrations**: Write raw SQL for both up and down migrations
- **Timestamp-Based Ordering**: Migrations are automatically sorted by creation time
- **Transaction Support**: Migrations run within transactions for safety
- **Migration Tracking**: Applied migrations are recorded in a database table
- **Dry Run Mode**: Validate migrations without applying them
- **Environment Variable Support**: Secure your database credentials
- **Rollback Support**: Easily revert migrations
- **Idempotent Migrations**: Well-written migrations can be run multiple times safely

## Installation

### Quick Install (macOS & Linux)

```bash
curl -sSL https://raw.githubusercontent.com/BolajiOlajide/kat/main/install.sh | bash
```

### From Pre-compiled Binaries

Download the appropriate binary for your operating system from the [releases page](https://github.com/BolajiOlajide/kat/releases).

### From Source

```bash
# Clone the repository
git clone https://github.com/BolajiOlajide/kat.git
cd kat

# Install
make install
```

For more installation options, see the [installation documentation](https://bolajiolajide.github.io/kat/install/).

## Quick Start

```bash
# Initialize a new Kat project
kat init

# Create a new migration
kat add create_users_table

# Edit the generated migration files in migrations/TIMESTAMP_create_users_table/

# Apply all pending migrations
kat up

# Roll back the most recent migration
kat down

# Test database connection
kat ping
```

## Usage

### Configuration

Kat uses a YAML configuration file (`kat.conf.yaml`) to specify:
- Database connection details
- Migration tracking table name
- Migration directory

Example configuration:

```yaml
migration:
  tablename: migrations
  directory: migrations
database:
  url: postgres://username:password@localhost:5432/mydatabase
```

### Commands

| Command | Description |
|---------|-------------|
| `kat init` | Initialize a new Kat project with configuration |
| `kat add NAME` | Create a new migration with the given name |
| `kat up` | Apply all pending migrations |
| `kat down [--count N]` | Roll back the most recent migration(s) |
| `kat ping` | Test database connectivity |
| `kat version` | Display the current version |
| `kat --help` | Show help for all commands |

For detailed usage instructions, see the [documentation](https://bolajiolajide.github.io/kat/).

## Migration Structure

Migrations are organized in a directory structure like this:

```
migrations/
  └─ 1679012345_create_users_table/
      ├─ up.sql      # SQL to apply the migration
      ├─ down.sql    # SQL to reverse the migration
      └─ metadata.yaml  # Migration metadata
```

## Documentation

Visit the [Kat documentation site](https://bolajiolajide.github.io/kat/) for detailed guides on:

- [Installation](https://bolajiolajide.github.io/kat/install/)
- [Initialization](https://bolajiolajide.github.io/kat/init/)
- [Configuration](https://bolajiolajide.github.io/kat/config/)
- [Database Connectivity](https://bolajiolajide.github.io/kat/ping/)
- [Working with Migrations](https://bolajiolajide.github.io/kat/migration/)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Kat is inspired by [Sourcegraph's internal CLI tooling](https://github.com/sourcegraph/sourcegraph-public-snapshot/tree/main/dev/sg).
