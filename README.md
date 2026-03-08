# Kat

<a href="https://www.producthunt.com/products/kat-3?embed=true&amp;utm_source=badge-featured&amp;utm_medium=badge&amp;utm_campaign=badge-kat-3" target="_blank" rel="noopener noreferrer"><img alt="Kat - Database migrations for the top 1% | Product Hunt" width="250" height="54" src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=1081944&amp;theme=neutral&amp;t=1771554914075"></a>

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/BolajiOlajide/kat/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/BolajiOlajide/kat)](https://goreportcard.com/report/github.com/BolajiOlajide/kat)
[![Go Reference](https://pkg.go.dev/badge/github.com/BolajiOlajide/kat.svg)](https://pkg.go.dev/github.com/BolajiOlajide/kat)
[![CI](https://github.com/BolajiOlajide/kat/actions/workflows/ci.yml/badge.svg)](https://github.com/BolajiOlajide/kat/actions/workflows/ci.yml)
[![Release](https://github.com/BolajiOlajide/kat/actions/workflows/release.yml/badge.svg)](https://github.com/BolajiOlajide/kat/actions/workflows/release.yml)
[![Docs](https://img.shields.io/badge/docs-kat.bolaji.de-blue)](https://kat.bolaji.de/)

**Graph-based SQL migrations.** Kat treats your schema as a DAG, not a linear log — enabling parallel development, explicit dependencies, and deterministic ordering. Supports PostgreSQL and SQLite.

![Kat Banner](doc/assets/images/layout/logo.png)

```text
Linear (traditional):     Graph-based (Kat):

001_users                 001_users ──┬─→ 003_posts
002_posts                             │
003_add_email             002_add_email ──┘

Order: 1→2→3              Order: 1→2→3 OR 1→3→2
(rigid)                   (flexible, dependency-aware)
```

## Install

```bash
# macOS & Linux
curl -sSL https://kat.bolaji.de/install | sudo bash

# From source
git clone https://github.com/BolajiOlajide/kat.git && cd kat && make install
```

Pre-compiled binaries are available on the [releases page](https://github.com/BolajiOlajide/kat/releases).

## Quick Start

```bash
kat init                        # Initialize project
kat add create_users_table      # Create a migration
kat add add_email_column        # Kat resolves the parent automatically
kat up                          # Apply all pending migrations
kat down --count 1              # Roll back last migration
kat ping                        # Test DB connection
kat export --file graph.dot     # Export dependency graph (DOT format)
```

### Migration Structure

```
migrations/
├── 1679012345_create_users_table/
│   ├── up.sql
│   ├── down.sql
│   └── metadata.yaml
└── 1679012398_add_email_column/
    ├── up.sql
    ├── down.sql
    └── metadata.yaml      # parents: [1679012345]
```

### Configuration (`kat.conf.yaml`)

```yaml
migration:
  tablename: migrations
  directory: migrations
database:
  driver: postgres
  url: postgres://user:pass@localhost:5432/mydb
  # or: url: ${DATABASE_URL}

  # driver: sqlite
  # path: ./kat.db
```

## Commands

| Command | Description |
|---------|-------------|
| `kat init` | Initialize a new project |
| `kat add NAME` | Create a new migration |
| `kat up [--count N]` | Apply pending migrations |
| `kat down [--count N]` | Roll back migrations |
| `kat ping` | Test DB connectivity |
| `kat export [--file F]` | Export migration graph (DOT format) |
| `kat version` | Display version |

## Go Library

```go
//go:embed migrations
var migrationsFS embed.FS

func main() {
    m, err := kat.New(kat.PostgresDriver, "postgres://user:pass@localhost:5432/db", migrationsFS, "migrations")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    m.Up(ctx, 0)   // apply all
    m.Down(ctx, 1)  // roll back one
}
```

## Comparison

| Feature | Kat | Flyway | Goose | Atlas |
|---------|-----|--------|-------|-------|
| Graph-based dependencies | ✅ | ❌ | ❌ | ⚠️ |
| Parallel development friendly | ✅ | ❌ | ❌ | ⚠️ |
| Raw SQL migrations | ✅ | ✅ | ✅ | ⚠️ |
| Go library + CLI | ✅ | ❌ | ✅ | ✅ |
| Migration visualization | ✅ | ❌ | ❌ | ✅ |

## Documentation

Full docs at [kat.bolaji.de](https://kat.bolaji.de/) — covers [installation](https://kat.bolaji.de/installation/), [configuration](https://kat.bolaji.de/config/), [migrations](https://kat.bolaji.de/migration/), [custom loggers](https://kat.bolaji.de/logger/), and more.

## Contributing

Contributions welcome! Fork, branch, commit, and open a PR.

## License

Apache License 2.0 — see [LICENSE](LICENSE).

## Acknowledgments

Inspired by [Sourcegraph's internal CLI tooling](https://github.com/sourcegraph/sourcegraph-public-snapshot/tree/main/dev/sg).
