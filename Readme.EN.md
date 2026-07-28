# zaplogs

A lightweight logging wrapper built on [uber-go/zap](https://github.com/uber-go/zap), providing out-of-the-box logging with console output, file rotation, and multi-logger management.

[中文](./Readme.md)

## Features

- **Zero-config startup** — Use global log functions without any initialization
- **Multi-logger management** — Create and manage multiple independent loggers by name
- **Log rotation** — Automatic log splitting and compression via [lumberjack](https://github.com/natefinch/lumberjack)
- **Flexible console format** — Customize output fields with a format string (Time/Level/Caller/Message)
- **Concurrency-safe** — All logger operations are thread-safe
- **High performance** — Built on zap with minimal overhead

## Installation

```bash
go get zaplogs
```

## Quick Start

### Zero-config Usage (Global Functions)

Call package-level functions directly without any initialization:

```go
package main

import "zaplogs"

func main() {
    zaplogs.Info("server started")
    zaplogs.Infof("listening on port: %d", 8080)
    zaplogs.Warn("low disk space")
    zaplogs.Errorf("connection failed: %v", err)
}
```

### Custom Configuration

```go
package main

import "zaplogs"

func main() {
    // Create config: level, log file path, console format
    config := zaplogs.NewLogConfig("debug", "./logs/app.log", "TLCM")
    
    // Initialize the default logger
    if err := zaplogs.InitDefaultLogger(config); err != nil {
        panic(err)
    }
    defer zaplogs.Sync()

    zaplogs.Debugf("debug info: %s", "hello")
    zaplogs.Infof("user %s logged in", "admin")
}
```

### Multiple Loggers

```go
package main

import "zaplogs"

func main() {
    // Create a business logger
    bizConfig := zaplogs.NewLogConfig("info", "./logs/biz.log", "LCM")
    bizLogger, err := zaplogs.CreateLogger("business", bizConfig)
    if err != nil {
        panic(err)
    }

    // Create an access logger
    accessConfig := zaplogs.NewLogConfig("info", "./logs/access.log", "TM")
    accessLogger, err := zaplogs.CreateLogger("access", accessConfig)
    if err != nil {
        panic(err)
    }

    bizLogger.Infof("order created: %s", orderID)
    accessLogger.Infof("GET /api/users 200 12ms")

    // Retrieve an existing logger
    logger, exists := zaplogs.GetLogger("business")
    if exists {
        logger.Info("retrieved logger by name")
    }

    // Close all loggers
    defer zaplogs.CloseAll()
}
```

## Configuration

### LogConfig Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Level` | string | `"info"` | Log level: `debug` / `info` / `warn` / `error` / `fatal` |
| `LogFile` | string | `""` | Log file path; empty string disables file output |
| `ConsoleFormat` | string | `"LCM"` | Console output format; `"off"` or empty disables console |
| `MaxSize` | int | `100` | Max size per log file (MB) |
| `MaxBackups` | int | `10` | Max number of backup files to retain |
| `MaxAge` | int | `30` | Max days to retain log files |
| `Compress` | bool | `true` | Whether to compress backup files |

### Console Format String

The format string is composed of the following characters:

| Char | Meaning | Example Output |
|------|---------|----------------|
| `T` | Time (ISO8601) | `2024-05-20T15:30:00.000+0800` |
| `L` | Log level | `INFO` / `ERROR` |
| `C` | Caller location | `main.go:20` |
| `M` | Log message | The actual message content |

Examples: `"TLCM"` outputs full info, `"LM"` outputs level and message only, `"off"` disables console.

## API Overview

### Global Functions (Using Default Logger)

```go
zaplogs.Debug(args ...interface{})
zaplogs.Debugf(template string, args ...interface{})
zaplogs.Info(args ...interface{})
zaplogs.Infof(template string, args ...interface{})
zaplogs.Warn(args ...interface{})
zaplogs.Warnf(template string, args ...interface{})
zaplogs.Error(args ...interface{})
zaplogs.Errorf(template string, args ...interface{})
zaplogs.Fatal(args ...interface{})
zaplogs.Fatalf(template string, args ...interface{})
zaplogs.Sync() error
```

### Logger Management

```go
zaplogs.CreateLogger(name string, config LogConfig) (*Logger, error)
zaplogs.GetLogger(name string) (*Logger, bool)
zaplogs.CloseAll() error
```

### Logger Instance Methods

```go
logger.Debug(args ...interface{})
logger.Debugf(template string, args ...interface{})
logger.Info(args ...interface{})
logger.Infof(template string, args ...interface{})
logger.Warn(args ...interface{})
logger.Warnf(template string, args ...interface{})
logger.Error(args ...interface{})
logger.Errorf(template string, args ...interface{})
logger.Fatal(args ...interface{})
logger.Fatalf(template string, args ...interface{})
logger.Sync() error
```

## Examples

See the [examples](./examples) directory for complete example code:

- [basic](./examples/basic) — Zero-config and custom configuration basics
- [multi-logger](./examples/multi-logger) — Multi-logger management
- [file-output](./examples/file-output) — File output with log rotation

## Dependencies

- [go.uber.org/zap](https://github.com/uber-go/zap) — High-performance structured logging
- [gopkg.in/natefinch/lumberjack.v2](https://github.com/natefinch/lumberjack) — Log rotation

## License

MIT
