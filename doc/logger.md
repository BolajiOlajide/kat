---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql,logger,logging
title: Logger
description: How to configure custom logging for Kat migrations
permalink: /logger
---

# Logger

Kat provides a customizable logger interface that allows you to control how migration messages are displayed or captured.

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

    m, err := kat.New(
        kat.WithLogger(logger),
        // other options...
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
    "github.com/BolajiOlajide/kat/internal/loggr"
)

func main() {
    m, err := kat.New(
        // other options...
    )
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

## Log Messages

During migration execution, Kat will log various messages including:

- Migration start/completion status
- Database connection details
- Migration file processing
- Error conditions
- Debug information about migration dependencies
