---
# Page settings
layout: default
keywords: kat,postgres,sqlite,database,cli,migrations,sql
title: Configuration
description: |
    Kat is a database migration tool for PostgreSQL and SQLite. It allows you run your migrations with raw SQL files.
comments: false
permalink: /config/
page_nav:
    prev:
        content: Initialization
        url: '/init'
    next:
        content: Editor Support
        url: '/editor-support'
---

# Understanding Kat Configuration

After [initializing](/init/) your project with `kat init`, a configuration file (`kat.conf.yaml`) is created for you. This guide explains how each configuration option works and how they enable successful database migrations.

## Configuration File Overview

The `kat.conf.yaml` file contains all the settings Kat needs to connect to your database and manage migrations. You can specify a different configuration file using the `--config` flag:

```bash
kat --config /path/to/your/config.yaml <command>
```

## Configuration Structure

Your configuration file has two main sections:
- `migration`: Settings related to migration files and tracking
- `database`: Connection details for your database

Here's the basic configuration structure:

```yaml
migration:
  tablename: migrations
  directory: migrations
database:
  driver: postgres
  url: postgres://username:password@hostname:5432/dbname
```

## Understanding Migration Configuration

The `migration` section defines how Kat manages and tracks your migrations:

| Option | Description | Default | Required |
|--------|-------------|---------|----------|
| `tablename` | Name of the table where Kat tracks applied migrations | `migrations` | No |
| `directory` | Directory where your SQL migration files are stored | `migrations` | No |

### How Migration Tracking Works

Kat creates a table (specified by `tablename`) in your database to track which migrations have been applied. This ensures migrations are only applied once and enables features like:

- Tracking migration history
- Applying pending migrations
- Rolling back migrations

The `directory` option tells Kat where to find your migration files. Each migration consists of "up" and "down" SQL files.

### Migration Configuration Example

```yaml
migration:
  tablename: migration_logs  # Custom table name for tracking
  directory: /path/to/migrations  # Path to migration files
  dryRun: false  # Actually apply migrations
```

## Database Driver

The `driver` field specifies which database backend to use. If omitted, it defaults to `postgres` for backward compatibility.

| Driver | Value | Description |
|--------|-------|-------------|
| PostgreSQL | `postgres` or `postgresql` | PostgreSQL database (default) |
| SQLite | `sqlite` or `sqlite3` | SQLite database (no CGO required) |

### PostgreSQL Configuration

```yaml
database:
  driver: postgres
  url: postgres://username:password@hostname:5432/dbname
```

### SQLite Configuration

For SQLite, use the `path` field instead of `url`:

```yaml
database:
  driver: sqlite
  path: file:///path/to/your/database.db
```

> 💡 SQLite uses a pure Go implementation via `modernc.org/sqlite` — no CGO or external C libraries are required.

## Understanding Database Configuration

For PostgreSQL, Kat offers two ways to configure your database connection:

### 1. Using a Connection URL

The simplest approach is to provide a PostgreSQL connection URL:

```yaml
database:
  url: postgres://username:password@hostname:5432/dbname?sslmode=disable
```

The URL format follows the standard PostgreSQL connection string format:
`postgres://[username]:[password]@[hostname]:[port]/[dbname]?[params]`

### 2. Using Individual Parameters

You can also specify connection details individually:

```yaml
database:
  user: username
  password: password
  host: localhost
  port: 5432
  name: dbname
  sslmode: disable
```

| Option | Description | Default | Required if URL not provided |
|--------|-------------|---------|----------|
| `user` | Database username | - | Yes |
| `password` | Database password | - | Yes |
| `host` | Database hostname or IP | - | Yes |
| `port` | Database port | `5432` | No |
| `name` | Database name | - | Yes |
| `sslmode` | SSL mode (disable, require, etc.) | `disable` | No |

### How Database Connection Works

Kat establishes a connection to your database using the provided credentials. This connection is used to:

1. Create/read the migration tracking table
2. Execute migration SQL scripts
3. Manage transactions during migrations

## Securing Database Credentials

Kat supports environment variables in the configuration file. Use `${VARIABLE_NAME}` syntax to reference environment variables:

```yaml
database:
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  host: ${DB_HOST}
  name: ${DB_NAME}
```

This allows you to keep sensitive information out of your configuration file and use different credentials across environments.

## Database Connection Tuning

Kat supports configurable connection timeouts and pool settings in `kat.conf.yaml`. These settings are optional — sensible defaults are used when not specified.

| Option | Description | Default |
|--------|-------------|---------|
| `connect_timeout` | Timeout for establishing connections (Go duration string) | `10s` |
| `statement_timeout` | Timeout for individual SQL statements | Disabled |
| `max_open_conns` | Maximum number of open connections | `2` (PostgreSQL), `1` (SQLite) |
| `max_idle_conns` | Maximum number of idle connections | `2` (PostgreSQL), `1` (SQLite) |
| `conn_max_lifetime` | Maximum connection lifetime (Go duration string) | `5m` (PostgreSQL), `2m` (SQLite) |
| `default_timeout` | Default timeout for operations without explicit deadline | Disabled |

### Example

```yaml
database:
  driver: postgres
  url: postgres://user:pass@localhost:5432/myapp
  connect_timeout: 5s
  statement_timeout: 5m
  max_open_conns: 20
  max_idle_conns: 10
  conn_max_lifetime: 1h
```

> ⚠️ For SQLite, `max_open_conns` is always enforced as `1` to prevent "database is locked" errors, regardless of the configured value.

## Configuration Examples for Common Scenarios

### Basic Local Development

```yaml
migration:
  tablename: migrations
  directory: migrations
database:
  user: postgres
  password: postgres
  host: localhost
  port: 5432
  name: myapp_development
  sslmode: disable
```

### Production with Environment Variables

```yaml
migration:
  tablename: migrations
  directory: /app/migrations
database:
  url: postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require
```

### CI/CD Pipeline

```yaml
migration:
  tablename: migrations
  directory: migrations
  dryRun: true  # Verify migrations without applying them
database:
  url: postgres://ci_user:${CI_DB_PASSWORD}@db-host/myapp_test
```

### SQLite Local Development

```yaml
migration:
  tablename: migrations
  directory: migrations
database:
  driver: sqlite
  path: file:///path/to/myapp.db
```

## Troubleshooting Configuration Issues

If you encounter issues with your configuration:

1. Verify your database connection details
2. Check that your migration directory exists and contains SQL files
3. Ensure your database user has sufficient permissions
4. Try running with the `--verbose` flag for more detailed logs

## Next Steps

Now that you understand how to configure Kat, you're ready to [create and run migrations](/migration/) to manage your database schema.