# zaplogs

A lightweight logging wrapper built on [uber-go/zap](https://github.com/uber-go/zap), providing out-of-the-box logging with console output, file rotation, and multi-logger management.

[中文](./Readme.md)

## Features

- **Zero-config startup** — Use global log functions without any initialization
- **Multi-logger management** — Create and manage multiple independent loggers by name
- **Log rotation** — Automatic log splitting and compression via [lumberjack](https://github.com/natefinch/lumberjack)
- **Flexible console format** — Three console output modes: mask (`TLCM` style), text (zap console encoder), and JSON; switch or disable anytime
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
    // Use a LogConfig struct literal (recommended for new code)
    config := zaplogs.LogConfig{
        Level:         "debug",           // debug / info / warn / error / fatal
        ConsoleFormat: "text",            // console: "" / "text" / "json" / "off"
        LogFilePath:   "./logs/app.log",  // file output path; empty disables file output
        LogFileFormat: "json",            // file format: "" defaults to json / "text" / "json" / "off"
        MaxSize:       100,               // max MB per file
        MaxBackups:    10,                // max backup files
        MaxAge:        30,                // retention in days
        Compress:      true,              // compress rotated files
    }

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
    // Business logger: console uses mask format (Level + Caller + Message), JSON file output
    bizConfig := zaplogs.LogConfig{
        Level:         "info",
        ConsoleFormat: "LCM",                // mask format (non-standard value = mask string): Level + Caller + Message
        LogFilePath:   "./logs/biz.log",
        LogFileFormat: "json",
        MaxSize:       100,
        MaxBackups:    10,
        MaxAge:        30,
        Compress:      true,
    }
    bizLogger, err := zaplogs.CreateLogger("business", bizConfig)
    if err != nil {
        panic(err)
    }

    // Access logger: console uses JSON format, also writes to file
    accessConfig := zaplogs.LogConfig{
        Level:         "info",
        ConsoleFormat: "json",               // slog json style
        LogFilePath:   "./logs/access.log",
        LogFileFormat: "json",
        MaxSize:       50,
        MaxBackups:    5,
        MaxAge:        7,
        Compress:      true,
    }
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
| `ConsoleFormat` | string | `""` | Console (stdout) format (see "Console Format" section). `""` defaults to mask `"LCM"`; any non-standard value (e.g. `"TLCM"`) is treated as a mask string (legacy compat). |
| `LogFileFormat` | string | `""` | File output format, same value options as `ConsoleFormat`. `""` defaults to `"json"`; non-standard values treated as mask strings. |
| `LogFilePath` | string | `""` | Log file path; empty string disables file output |
| `MaxSize` | int | `100` | Max size per log file (MB) before rotation |
| `MaxBackups` | int | `10` | Max backup files retained (`NewLogConfigEmpty` uses `3`) |
| `MaxAge` | int | `30` | Max days to retain old log files |
| `Compress` | bool | `true` | Whether to compress rotated files |

### Console Format

`ConsoleFormat` accepts the following values:

| Value | Description |
|-------|-------------|
| `""` | Default; falls back to the mask format `"LCM"` |
| `"text"` | zap built-in console encoder: `<time>\t<level>\t<caller>\t<message>\t{"field":"value", ...}`. The human-readable prefix is tab-separated; any structured fields are appended as a JSON object. Example: `2024-05-20T15:30:00.000+0800\tINFO\tmain.go:20\tuser logged in\t{"uid":42}` |
| `"json"` | zap JSON encoder (slog json style): `{"level":"info","ts":"...","caller":"main.go:20","msg":"..."}` |
| `"off"` | Disable console output |
| any other non-empty value | Treated as a mask string (legacy compatibility) |

Mask format characters (used when `ConsoleFormat` / `LogFileFormat` is set to a non-standard value):

| Char | Meaning | Example Output |
|------|---------|----------------|
| `T` | Time (ISO8601) | `2024-05-20T15:30:00.000+0800` |
| `L` | Log level | `INFO` / `ERROR` |
| `C` | Caller location | `main.go:20` |
| `M` | Log message | The actual message content |

Examples: `"TLCM"` outputs full info, `"LM"` outputs level and message only. Console output is written to stdout.

### Configuration Functions

```go
zaplogs.NewLogConfig(level, logFile, consoleFormat string) LogConfig
zaplogs.NewLogConfigEmpty() LogConfig
```

`NewLogConfig` is the **legacy convenience constructor**: its 3rd argument `consoleFormat` is written directly to `ConsoleFormat` (passing `"LCM"` is equivalent to `ConsoleFormat: "LCM"`, treated as a mask string) for backward compatibility with older versions. `NewLogConfig` fills defaults: `MaxSize=100`, `MaxBackups=10`, `MaxAge=30`, `Compress=true`. `NewLogConfigEmpty` returns all defaults (`MaxBackups` defaults to `3`, others identical to `NewLogConfig`).

New code is encouraged to use a `LogConfig` struct literal or load configuration from YAML (see below) for full control over all fields.

### YAML Configuration

Every field of `LogConfig` carries a `yaml` tag, so it can be deserialized directly with any standard YAML library. **The library itself does not provide a `LoadConfig` function** — callers read and parse the YAML themselves, then pass the resulting `LogConfig` to `InitDefaultLogger` / `CreateLogger`.

Example `config.yaml`:

```yaml
level: info            # debug / info / warn / error / fatal
console_format: text   # "" / text / json / off ("" defaults to mask "LCM"); non-standard values like "TLCM" are treated as mask strings
log_file_format: json  # "" defaults to json / text / json / off; non-standard values treated as mask strings
log_file_path: ./logs/app.log
max_size: 100          # max MB per file
max_backups: 10        # max backup files retained
max_age: 30            # retention in days
compress: true          # compress rotated files
```

Load and initialize (the caller must import a YAML library, e.g. `gopkg.in/yaml.v3`):

```go
package main

import (
    "os"

    "gopkg.in/yaml.v3" // caller's choice of YAML library
    "zaplogs"
)

func main() {
    data, err := os.ReadFile("config.yaml")
    if err != nil {
        panic(err)
    }

    var cfg zaplogs.LogConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        panic(err)
    }

    if err := zaplogs.InitDefaultLogger(cfg); err != nil {
        panic(err)
    }
    defer zaplogs.Sync()
}
```

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
zaplogs.DefaultLogger() *Logger
zaplogs.DefaultZapLogger() *zap.Logger
```

> `DefaultLogger()` returns the default `*Logger` instance, which can be assigned to custom interface types or used for dependency injection; `DefaultZapLogger()` returns the underlying `*zap.Logger` for structured-logging consumers.

### Logger Management

```go
zaplogs.CreateLogger(name string, config LogConfig) (*Logger, error)
zaplogs.GetLogger(name string) (*Logger, bool)
zaplogs.CloseAll() error
```

> `CloseAll()` first flushes all logger buffers, then closes the underlying log files (releasing file handles), and resets the default logger — subsequent global function calls will automatically re-initialize the default logger.

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
logger.ZapLogger() *zap.Logger
logger.SugaredLogger() *zap.SugaredLogger
logger.Close() error
```

> `ZapLogger()` / `SugaredLogger()` return the underlying raw zap objects for third-party libraries that need structured fields or the sugared API; `Close()` closes the logger's underlying file — make sure no concurrent writes are in progress before calling it.

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
