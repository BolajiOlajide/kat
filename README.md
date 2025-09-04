# Kat

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/BolajiOlajide/kat/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/BolajiOlajide/kat)](https://goreportcard.com/report/github.com/BolajiOlajide/kat)
[![Go Reference](https://pkg.go.dev/badge/github.com/BolajiOlajide/kat.svg)](https://pkg.go.dev/github.com/BolajiOlajide/kat)
[![CI](https://github.com/BolajiOlajide/kat/actions/workflows/ci.yml/badge.svg)](https://github.com/BolajiOlajide/kat/actions/workflows/ci.yml)
[![Release](https://github.com/BolajiOlajide/kat/actions/workflows/release.yml/badge.svg)](https://github.com/BolajiOlajide/kat/actions/workflows/release.yml)
[![Docs](https://img.shields.io/badge/docs-kat.bolaji.de-blue)](https://kat.bolaji.de/)

<details>
<summary>📑 Table of Contents</summary>

- [Why Graph-Based Migrations?](#why-graph-based-migrations)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Comparison with Other Tools](#comparison-with-other-tools)
- [Architecture Overview](#architecture-overview)
- [Go Library Usage](#go-library-usage)
- [Best Practices](#best-practices)
- [Limitations](#limitations)
- [Quick Reference](#quick-reference)
- [Documentation](#documentation)

</details>

Kat is (probably?) the first open-source PostgreSQL migration tool that treats your schema as a Directed Acyclic Graph, not a linear log. It enables topological sort migrations with explicit dependencies, parallel development workflows, and deterministic ordering.

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

## Why Graph-Based Migrations?

Traditional migration tools force you into a linear sequence where every developer must coordinate their schema changes. Kat's graph-based approach solves this by allowing migrations to declare explicit parent dependencies, creating a Directed Acyclic Graph (DAG) that determines execution order.

**Benefits of the DAG approach:**
- **Parallel Development**: Multiple developers can create feature migrations simultaneously without conflicts
- **Deterministic Ordering**: Kat computes the optimal execution order based on dependencies, not timestamps
- **Safe Branching**: Feature branches can include their own migrations that merge cleanly
- **Complex Dependencies**: A migration can depend on multiple parents, enabling sophisticated schema evolution

```text
Linear (traditional):     Graph-based (Kat):

001_users                 001_users ──┬─→ 003_posts
002_posts                             │
003_add_email             002_add_email ──┘

Order: 1→2→3              Order: 1→2→3 OR 1→3→2
(rigid)                   (flexible, dependency-aware)
```

This means you can create migrations for different features in parallel, and Kat will figure out the correct order when you run `kat up`.

## Installation

### Quick Install (macOS & Linux)

```bash
curl -sSL https://kat.bolaji.de/install | sudo bash
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

Here's a realistic example showing how Kat's graph-based system handles parallel development:

```bash
# Initialize a new Kat project
kat init

# Create foundation migration
kat add create_users_table

# Developer A: Add email feature
kat add add_email_column --parent create_users_table

# Developer B: Add posts feature (can work in parallel)
kat add create_posts_table --parent create_users_table

# Developer C: Add full-text search (depends on both email and posts)
kat add add_full_text_search --parent add_email_column --parent create_posts_table

# Visualize the dependency graph
kat export --file graph.dot
dot -Tpng graph.dot -o migrations.png  # Requires Graphviz

# Apply all migrations - Kat determines the correct order automatically
kat up

# Test database connection
kat ping

# Roll back specific number of migrations
kat down --count 2
```

### Example Migration Structure

Each migration is a directory containing SQL files and metadata:

```
migrations/
├── 1679012345_create_users_table/
│   ├── up.sql
│   ├── down.sql
│   └── metadata.yaml
├── 1679012398_add_email_column/
│   ├── up.sql
│   ├── down.sql
│   └── metadata.yaml      # parents: [1679012345]
└── 1679012401_create_posts_table/
    ├── up.sql
    ├── down.sql
    └── metadata.yaml          # parents: [1679012345]
```

**metadata.yaml example:**
```yaml
id: 1679012398
name: add_email_column
description: Add email column to users table
parents:
  - 1679012345  # create_users_table
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

## Comparison with Other Tools

| Feature | Kat | Flyway | Goose | Atlas |
|---------|-----|---------|-------|-------|
| Graph-based dependencies | ✅ | ❌ | ❌ | ⚠️ |
| Parallel development friendly | ✅ | ❌ | ❌ | ⚠️ |
| Raw SQL migrations | ✅ | ✅ | ✅ | ⚠️ |
| Go library + CLI | ✅ | ❌ | ✅ | ✅ |
| Transaction per migration | ✅ | ✅ | ✅ | ✅ |
| Rollback support | ✅ | ✅ | ✅ | ✅ |
| Migration visualization | ✅ | ❌ | ❌ | ✅ |

## Architecture Overview

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│ CLI Command │ -> │   Migration  │ -> │    Graph    │ -> │   Runner     │
│   (cmd/)    │    │ (discovery)  │    │  (DAG ops)  │    │ (execution)  │
└─────────────┘    └──────────────┘    └─────────────┘    └──────────────┘
                           │                    │                 │
                           v                    v                 v
                   ┌──────────────┐    ┌─────────────┐    ┌──────────────┐
                   │  File System │    │ Topological │    │   Database   │
                   │   Scanner    │    │    Sort     │    │  Operations  │
                   └──────────────┘    └─────────────┘    └──────────────┘
```

**Flow**: Discovery → Graph Construction → Topological Ordering → Transactional Execution

## Go Library Usage

Kat can also be used as a Go library in your applications:

```go
package main

import (
    "context"
    "embed"
    "log"
    "time"

    "github.com/BolajiOlajide/kat"
)

//go:embed migrations
var migrationsFS embed.FS

func main() {
    // Basic usage with embedded migrations
    m, err := kat.New("postgres://user:pass@localhost:5432/db", migrationsFS, "migrations")
    if err != nil {
        log.Fatal(err)
    }

    // Apply all pending migrations with cancellation
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    err = m.Up(ctx, 0) // 0 = apply all pending
    if err != nil {
        log.Fatal(err)
    }

    // Roll back the most recent migration
    err = m.Down(context.Background(), 1)
    if err != nil {
        log.Fatal(err)
    }

    // Advanced usage with custom logger and existing DB connection
    connStr := "postgres://user:pass@localhost:5432/db"
    db, _ := sql.Open("postgres", connStr)
    m, err = kat.New("", migrationsFS, "schema_migrations",
        kat.WithLogger(customLogger),
        kat.WithSqlDB(db), // Reuse existing connection
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

For more details on custom logging, see the [logger documentation](https://kat.bolaji.de/logger/).

## Best Practices

### Writing Idempotent Migrations
```sql
-- up.sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE
);

-- down.sql
DROP TABLE IF EXISTS users;
```

### Branch Workflow
1. Create feature branch: `git checkout -b feature/user-profiles`
2. Add migrations with proper parents: `kat add add_profile_table --parent create_users_table`
3. Test locally: `kat up --dry-run`
4. Merge: Kat automatically handles dependency ordering

### CI/CD Integration
```yaml
# .github/workflows/migrations.yml
- name: Validate migrations
  run: |
    kat ping
    kat up --dry-run
```

## Limitations

- **PostgreSQL only**: Requires PostgreSQL ≥9.5 for transactional DDL support
- **Acyclic constraint**: Migration graph must remain cycle-free
- **No concurrent execution protection**: Multiple Kat instances can race (advisory locks [planned](https://github.com/BolajiOlajide/kat/issues))
- **No content verification**: Applied migrations aren't checksummed for drift detection ([planned](https://github.com/BolajiOlajide/kat/issues))

**Compatibility**: Tested with Go 1.20+ (tested on 1.23), PostgreSQL 12-16. Supported OS: Linux, macOS, Windows (amd64/arm64)

## Quick Reference

| Command | Purpose | Common Flags |
|---------|---------|--------------|
| `kat add NAME --parent <id>` | Create migration with dependency | `--parent, -p` |
| `kat up --count 3` | Apply next 3 migrations | `--dry-run, --verbose` |
| `kat down --count 2` | Roll back 2 migrations | `--force` |
| `kat export --file graph.dot` | Export dependency graph | `--format json` |
| `kat ping` | Test database connectivity | |

➡️ *Need help?* Visit [GitHub Discussions](https://github.com/BolajiOlajide/kat/discussions) for questions and [GitHub Issues](https://github.com/BolajiOlajide/kat/issues) for bug reports.

## Documentation

Visit the [Kat documentation site](https://kat.bolaji.de/) for detailed guides:

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
