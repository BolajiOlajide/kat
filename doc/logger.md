---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql,logger,logging
title: Logger
description: How to configure custom logging for Kat migrations
permalink: /logger
---

# Custom Logging

**You don't have to do this** - Kat already prints colorized logs by default.

Kat provides a customizable logger interface that allows you to control how migration messages are displayed or captured. This is useful for integrating with existing logging systems or customizing output format.

## Default Logger

By default, Kat uses a built-in logger that prints messages to stdout with colored output:

- **Debug**: Suggestion-style output (cyan)
- **Info**: Info-style output (blue)
- **Warn**: Warning-style output (yellow)
- **Error**: Error-style output (red)

## Custom Logger

You can provide your own logger implementation by implementing the `Logger` interface:

```go
type Logger interface {
    Debug(msg string)
    Info(msg string)
    Warn(msg string)
    Error(msg string)
}
```

## Usage

### With Custom Logger

```go
package main

import "github.com/BolajiOlajide/kat"

type MyLogger struct{}

func (l *MyLogger) Debug(msg string) {
    // Your debug logging implementation
}

func (l *MyLogger) Info(msg string) {
    // Your info logging implementation
}

func (l *MyLogger) Warn(msg string) {
    // Your warn logging implementation
}

func (l *MyLogger) Error(msg string) {
    // Your error logging implementation
}

func main() {
    var logger kat.Logger
    logger = &MyLogger{}

    m, err := kat.New(kat.PostgresDriver, "postgres://user:pass@localhost:5432/db", fsys, "migrations",
        kat.WithLogger(logger),
    )
    if err != nil {
        // handle error
    }

    // Use migration instance
}
```

### With Default Logger

```go
package main

import (
    "github.com/BolajiOlajide/kat"
)

func main() {
    m, err := kat.New(kat.PostgresDriver, "postgres://user:pass@localhost:5432/db", fsys, "migrations")
    if err != nil {
        // handle error
    }

    // Use migration instance
}
```

## Integration with Other Logging Libraries

You can easily integrate Kat with popular logging libraries by creating adapter implementations:

```go
// Example with logrus
type LogrusAdapter struct {
    logger *logrus.Logger
}

func (l *LogrusAdapter) Debug(msg string) {
    l.logger.Debug(msg)
}

func (l *LogrusAdapter) Info(msg string) {
    l.logger.Info(msg)
}

func (l *LogrusAdapter) Warn(msg string) {
    l.logger.Warn(msg)
}

func (l *LogrusAdapter) Error(msg string) {
    l.logger.Error(msg)
}
```

### With Existing Database Connection

If you already have a `*sql.DB` instance, you can use `NewWithDB`:

```go
package main

import (
    "database/sql"
    "github.com/BolajiOlajide/kat"
)

func main() {
    // Existing database connection
    db, err := sql.Open("pgx", "postgres://user:pass@localhost:5432/db")
    if err != nil {
        // handle error
    }
    defer db.Close()

    m, err := kat.NewWithDB(kat.PostgresDriver, db, fsys, "migrations")
    if err != nil {
        // handle error
    }

    // Use migration instance
}
```

## Migration Options

Kat supports several configuration options through the `MigrationOption` type:

### WithLogger

Provides a custom logger implementation:

```go
m, err := kat.New(kat.PostgresDriver, connStr, fsys, "migrations",
    kat.WithLogger(customLogger),
)
```

### NewWithDB

Use `NewWithDB` to provide an existing `*sql.DB` connection instead of letting Kat create one:

```go
m, err := kat.NewWithDB(kat.PostgresDriver, existingDB, fsys, "migrations")
```

**Note:** When using `NewWithDB`, the caller is responsible for managing the connection lifecycle. Database configuration options (`WithDBConfig`, `WithConnectTimeout`, `WithPoolLimits`) are not supported — configure the `*sql.DB` directly. For SQLite, Kat automatically enforces `MaxOpenConns=1` to avoid "database is locked" errors.

### WithDBConfig

Configures custom database connection settings (only with `New`, not `NewWithDB`):

```go
config := kat.DBConfig{
    ConnectTimeout:   5 * time.Second,
    StatementTimeout: 5 * time.Minute,
    MaxOpenConns:     20,
    MaxIdleConns:     10,
    ConnMaxLifetime:  1 * time.Hour,
    DefaultTimeout:   60 * time.Second,
}
m, err := kat.New(kat.PostgresDriver, connStr, fsys, "migrations",
    kat.WithDBConfig(config),
)
```

### WithConnectTimeout

Convenience function to configure just the connection timeout:

```go
m, err := kat.New(kat.PostgresDriver, connStr, fsys, "migrations",
    kat.WithConnectTimeout(5 * time.Second),
)
```

### WithPoolLimits

Configures connection pool limits:

```go
m, err := kat.New(kat.PostgresDriver, connStr, fsys, "migrations",
    kat.WithPoolLimits(20, 10, 1*time.Hour),
)
```

## Log Messages

During migration execution, Kat will log various messages including:

- Migration start/completion status
- Database connection details
- Migration file processing
- Error conditions
- Debug information about migration dependencies
