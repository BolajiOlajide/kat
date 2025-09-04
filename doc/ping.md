---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql
title: Database Connectivity
description: |
    Check your PostgreSQL database connection with Kat's ping command.
comments: false
permalink: /ping/
page_nav:
    prev:
        content: Configuration
        url: '/config'
    next: 
        content: Migration
        url: '/migration'
---

# Testing Database Connectivity

**TL;DR**: `kat ping` verifies credentials and waits for database readiness.

Before running migrations, it's important to verify that Kat can connect to your PostgreSQL database. The `ping` command provides a simple way to test your database connection with optional retry capabilities for unstable connections.

## Basic Usage

To verify your database connection using the settings in your `kat.conf.yaml` file:

```bash
kat ping
```

When successful, you'll see:

```
Attempting to ping database
Successfully connected to database!
```

If the connection fails, Kat will provide detailed error information to help you troubleshoot.

## Retry Capabilities

The `ping` command includes built-in retry functionality, which is particularly useful when:

- Working with cloud databases that may have intermittent connectivity
- Testing connections to databases that are still initializing
- Verifying connectivity in container or orchestration environments

### Configuring Retries

The ping command accepts two retry-related parameters:

| Parameter | Description | Default | Range |
|-----------|-------------|---------|-------|
| `--retry-count` or `-r` | Number of retry attempts | 3 | 0-7 |
| `--retry-delay` or `-rd` | Initial delay in milliseconds | 500 | 100-3000 |

When retries are enabled, Kat uses an exponential backoff strategy, doubling the delay after each failed attempt.

### Examples

**Basic ping with default retries:**
```bash
kat ping
```

**Ping with 5 retries, starting with a 1-second delay:**
```bash
kat ping --retry-count 5 --retry-delay 1000
```

**Ping with no retries:**
```bash
kat ping --retry-count 0
```

## Using with Custom Configuration

To test connectivity with a specific configuration file:

```bash
kat ping --config /path/to/custom-config.yaml
```

## Environment Variables

You can also control retry behavior using environment variables:

| Environment Variable | Description | Equivalent Flag |
|----------------------|-------------|----------------|
| `KAT_RETRY_COUNT` | Number of retry attempts | `--retry-count` |
| `KAT_RETRY_DELAY` | Initial delay in milliseconds | `--retry-delay` |
| `KAT_CONFIG_FILE` | Custom configuration file path | `--config` |

Example:
```bash
KAT_RETRY_COUNT=5 KAT_RETRY_DELAY=1000 kat ping
```

## Docker Compose Wait-for-Database

Common pattern for waiting for PostgreSQL to be ready in Docker environments:

```bash
# docker-compose.yml service
healthcheck:
  test: ["CMD-SHELL", "kat ping --retry-count 5 --retry-delay 1000"]
  interval: 10s
  timeout: 5s
  retries: 3
```

<details>
<summary>📋 PostgreSQL Error Codes (Click to expand)</summary>

When a connection fails, Kat identifies and handles common PostgreSQL connection errors. The following error types are recognized as transient and will be retried:

- Connection exceptions (code 08003)
- Connection failures (code 08006)  
- Client unable to establish connection (code 08001)
- Server rejected connection (code 08004)
- Connection failures during transaction (code 08007)
- Server shutdowns (codes 57P01, 57P02, 57P03)
- Too many connections (codes 53300, 53301)

For non-transient errors (like authentication failures or invalid hostnames), Kat will immediately report the error without retrying.

</details>

## Usage in Scripts and CI/CD

The `ping` command is particularly useful in scripts and CI/CD pipelines to:

1. Verify database availability before running migrations
2. Wait for a database to become available before proceeding
3. Validate configuration without attempting migrations

Example script usage:
```bash
# Wait for database to be available before running migrations
if kat ping --retry-count 5 --retry-delay 2000; then
    echo "Database is available, running migrations..."
    kat up
else
    echo "Database connection failed after retries"
    exit 1
fi
```

## Troubleshooting

If you're having trouble connecting to your database with the ping command:

1. Verify your database connection details in `kat.conf.yaml`
2. Check that your PostgreSQL server is running and accepting connections
3. Ensure network connectivity between your environment and the database server
4. Verify that your database user has appropriate connection privileges
5. Check for any firewall or security group restrictions

For more advanced debugging, you can use the `--verbose` flag to see additional details:

```bash
kat --verbose ping
```