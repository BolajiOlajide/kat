---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql
title: Initializing Kat
description: |
    Learn how to set up your project with Kat's initialization command.
comments: false
permalink: /init/
page_nav:
    prev:
        content: Installation
        url: '/install'
    next:
        content: Configuration
        url: '/config'
---

# Initializing a Project with Kat

The `kat init` command is your starting point for setting up a new project with Kat. This command creates a configuration file (`kat.conf.yaml`) that defines your database connection details and migration settings.

## Basic Usage

```bash
kat init
```

This creates a configuration file with default settings in your current working directory.

## Command Flags

Kat init supports various flags to customize your configuration:

| Flag | Short | Environment Variable | Default | Description |
|------|-------|---------------------|---------|-------------|
| `--tableName` | `-t` | `KAT_MIGRATION_TABLE_NAME` | `migrations` | The database table where migration records are stored |
| `--directory` | `-d` | `KAT_MIGRATION_DIRECTORY` | `migrations` | Directory to store migration files |
| `--databaseURL` | `-u` | `KAT_MIGRATION_DATABASE_URL` | `` | PostgreSQL connection string (if provided, individual connection params are ignored) |
| `--dbUser` | | `KAT_DB_USER` | `postgres` | Database username |
| `--dbPassword` | | `KAT_DB_PASSWORD` | `postgres` | Database password |
| `--dbName` | | `KAT_DB_NAME` | `myapp` | Database name |
| `--dbPort` | | `KAT_DB_PORT` | `5432` | Database port |
| `--dbHost` | | `KAT_DB_HOST` | `localhost` | Database host |
| `--dbSSLMode` | | `KAT_DB_SSL_MODE` | `disable` | SSL mode (disable, allow, prefer, require, verify-ca, verify-full) |

## Using Environment Variables in Configuration

Kat supports using environment variables in your configuration file. When the configuration is loaded, environment variables referenced with the `$` symbol are automatically expanded.

### Examples:

```yaml
# Using environment variables directly
database:
  - user: $DB_USER
  - password: $DB_PASSWORD
  - name: $DATABASE_NAME
  - port: 5432
  - host: localhost
```

You can also use the braces syntax for clarity:

```yaml
database:
  - user: ${DB_USER}
  - password: ${DB_PASSWORD}
  - name: ${DATABASE_NAME}
```

This feature allows you to:
- Keep sensitive information like passwords out of version control
- Easily configure different environments (development, staging, production)
- Follow the twelve-factor app methodology for configuration

## Examples

### Basic Initialization

Create a default configuration file:

```bash
kat init
```

### Custom Table and Directory

Specify a custom migration table and directory:

```bash
kat init --tableName schema_migrations --directory db/migrations
```

### Using Connection String

Initialize with a database connection URL:

```bash
kat init --databaseURL "postgres://user:password@localhost:5432/mydb?sslmode=disable"
```

### Using Environment Variables as Flags

```bash
# Set environment variables
export KAT_DB_USER="admin"
export KAT_DB_PASSWORD="secure_password"

# Run init using those environment variables
kat init
```

### Complete Custom Configuration

```bash
kat init \
  --tableName schema_history \
  --directory db/migrations \
  --dbUser admin \
  --dbPassword secure_password \
  --dbName production_db \
  --dbPort 5432 \
  --dbHost db.example.com \
  --dbSSLMode require
```

## Result

After running `kat init`, a configuration file named `kat.conf.yaml` will be created in your current directory. This file will contain your specified settings or default values where not specified.

## Next Steps

Once you have initialized your project with `kat init`, you can:

1. Create your first migration with [`kat add`](/migration/#adding-migrations)
2. Apply migrations with [`kat up`](/migration/#applying-migrations)
3. Roll back migrations with [`kat down`](/migration/#rolling-back-migrations)

---