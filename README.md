# Kat

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/BolajiOlajide/kat/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/BolajiOlajide/kat)](https://goreportcard.com/report/github.com/BolajiOlajide/kat)
[![CI](https://github.com/BolajiOlajide/kat/actions/workflows/ci.yml/badge.svg)](https://github.com/BolajiOlajide/kat/actions/workflows/ci.yml)
[![Release](https://github.com/BolajiOlajide/kat/actions/workflows/release.yml/badge.svg)](https://github.com/BolajiOlajide/kat/actions/workflows/release.yml)

Kat is a lightweight, powerful CLI tool for PostgreSQL database migrations. It allows you to manage your database schema using SQL files with a simple, intuitive workflow.

![Kat Banner](doc/assets/images/layout/logo.png)

## Features

- **Simple SQL Migrations**: Write raw SQL for both up and down migrations
- **Graph-Based Migration System**: Manages parent-child relationships between migrations using a directed acyclic graph
- **Explicit Dependencies**: Migrations can declare parent dependencies to ensure proper execution order
- **Transaction Support**: Migrations run within transactions for safety
- **Migration Tracking**: Applied migrations are recorded in a database table
- **Dry Run Mode**: Validate migrations without applying them
- **Environment Variable Support**: Secure your database credentials
- **Rollback Support**: Easily revert migrations
- **Idempotent Migrations**: Well-written migrations can be run multiple times safely
- **Custom Logger Support**: Configure custom logging for migrations
- **Go Library**: Use Kat programmatically in your Go applications

## Installation

### Quick Install (macOS & Linux)

```bash
curl -sSL https://kat.bolaji.de/install | bash
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

For more installation options, see the [installation documentation](https://kat.bolaji.de/install/).

## Quick Start

```bash
# Initialize a new Kat project
kat init

# Create a new migration
kat add create_users_table

# Create another migration with a parent dependency
kat add create_posts_table

# Edit the generated migration files in migrations/TIMESTAMP_create_users_table/

# Apply all pending migrations (ordered based on dependencies)
kat up

# Roll back the most recent migration
kat down

# Test database connection
kat ping

# Visualize your migration graph (requires Graphviz)
kat export --file migrations.dot
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
  # Alternatively, use environment variables for secure credential management:
  # url: ${DATABASE_URL}
  # Or specify individual connection parameters:
  # host: ${DB_HOST}
  # port: ${DB_PORT}
  # user: ${DB_USER}
  # password: ${DB_PASSWORD}
  # dbname: ${DB_NAME}
```

### Commands

| Command                        | Description |
|--------------------------------|-------------|
| `kat init`                     | Initialize a new Kat project with configuration |
| `kat add NAME`                 | Create a new migration with the given name |
| `kat up [--count / -n]`        | Apply all pending migrations |
| `kat down [--count / -n]`      | Roll back the most recent migration(s) |
| `kat ping`                     | Test database connectivity |
| `kat export [--file FILENAME]` | Export the migration graph in DOT format for visualization |
| `kat version`                  | Display the current version |
| `kat --help`                   | Show help for all commands |

For detailed usage instructions, see the [documentation](https://kat.bolaji.de/).

## Go Library Usage

Kat can also be used as a Go library in your applications:

```go
package main

import (
    "context"
    "embed"
    "log"

    "github.com/BolajiOlajide/kat"
)

//go:embed migrations
var migrationsFS embed.FS

func main() {
    // Basic usage
    m, err := kat.New("postgres://user:pass@localhost:5432/db", migrationsFS, "migrations")
    if err != nil {
        log.Fatal(err)
    }

    // Apply all pending migrations
    err = m.Up(context.Background(), 0)
    if err != nil {
        log.Fatal(err)
    }

    // With custom logger
    m, err = kat.New("postgres://user:pass@localhost:5432/db", migrationsFS, "migrations",
        kat.WithLogger(customLogger),
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

For more details on custom logging, see the [logger documentation](https://kat.bolaji.de/logger/).

## Migration Structure

Migrations are organized in a directory structure like this:

```
migrations/
  └─ 1679012345_create_users_table/
      ├─ up.sql      # SQL to apply the migration
      ├─ down.sql    # SQL to reverse the migration
      └─ metadata.yaml  # Migration metadata including dependencies
```

## Documentation

Visit the [Kat documentation site](https://kat.bolaji.de/) for detailed guides on:

- [Installation](https://kat.bolaji.de/installation/)
- [Initialization](https://kat.bolaji.de/init/)
- [Configuration](https://kat.bolaji.de/config/)
- [Database Connectivity](https://kat.bolaji.de/ping/)
- [Working with Migrations](https://kat.bolaji.de/migration/)
- [Custom Logger Configuration](https://kat.bolaji.de/logger/)

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
