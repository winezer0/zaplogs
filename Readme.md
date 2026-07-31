# zaplogs

一个基于 [uber-go/zap](https://github.com/uber-go/zap) 的轻量级日志封装库，提供开箱即用的日志功能，支持控制台输出、文件轮转、多日志器管理。

[English](./Readme.EN.md)

## 特性

- **零配置启动** — 无需任何初始化即可直接使用全局日志函数
- **多日志器管理** — 按名称创建和管理多个独立日志器
- **日志轮转** — 基于 [lumberjack](https://github.com/natefinch/lumberjack) 实现自动日志切割与压缩
- **灵活的控制台格式** — 支持掩码（`TLCM` 风格）/ key=value 文本 / JSON 三种输出模式，可随时切换或关闭
- **并发安全** — 所有日志器操作均为线程安全
- **高性能** — 底层基于 zap，保持极低开销

## 安装

```bash
go get zaplogs
```

## 快速开始

### 零配置使用（全局函数）

无需任何初始化，直接调用包级函数即可输出日志：

```go
package main

import "zaplogs"

func main() {
    zaplogs.Info("服务启动")
    zaplogs.Infof("监听端口: %d", 8080)
    zaplogs.Warn("磁盘空间不足")
    zaplogs.Errorf("连接失败: %v", err)
}
```

### 自定义配置

```go
package main

import "zaplogs"

func main() {
    // 使用 LogConfig 结构体直接配置（推荐方式）
    config := zaplogs.LogConfig{
        Level:         "debug",           // debug / info / warn / error / fatal
        ConsoleFormat: "text",            // 控制台: "" / "text" / "json" / "off"
        LogFilePath:   "./logs/app.log",  // 文件输出路径，空串表示不输出文件
        LogFileFormat: "json",            // 文件格式: "" 默认 json / "text" / "json" / "off"
        MaxSize:       100,               // 单文件最大 100MB
        MaxBackups:    10,                // 最多保留 10 个备份
        MaxAge:        30,                // 保留 30 天
        Compress:      true,              // 压缩备份文件
    }

    // 初始化默认日志器
    if err := zaplogs.InitDefaultLogger(config); err != nil {
        panic(err)
    }
    defer zaplogs.Sync()

    zaplogs.Debugf("调试信息: %s", "hello")
    zaplogs.Infof("用户 %s 登录成功", "admin")
}
```

### 多日志器

```go
package main

import "zaplogs"

func main() {
    // 业务日志器：控制台使用掩码格式（级别 + 调用者 + 消息），写入 JSON 文件
    bizConfig := zaplogs.LogConfig{
        Level:         "info",
        ConsoleFormat: "LCM",                // 掩码格式（非标准值 = 掩码串）：级别 + 调用者 + 消息
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

    // 访问日志器：控制台使用 JSON 格式，同时写入文件
    accessConfig := zaplogs.LogConfig{
        Level:         "info",
        ConsoleFormat: "json",               // slog json 风格
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

    bizLogger.Infof("订单创建成功: %s", orderID)
    accessLogger.Infof("GET /api/users 200 12ms")

    // 获取已有日志器
    logger, exists := zaplogs.GetLogger("business")
    if exists {
        logger.Info("通过名称获取日志器")
    }

    // 关闭所有日志器
    defer zaplogs.CloseAll()
}
```

## 配置说明

### LogConfig 字段

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Level` | string | `"info"` | 日志级别: `debug` / `info` / `warn` / `error` / `fatal` |
| `ConsoleFormat` | string | `""` | 控制台输出格式（见"控制台格式"节）。`""` 默认回退到掩码 `"LCM"`；非标准值（如 `"TLCM"`）视为掩码串（旧版兼容） |
| `LogFileFormat` | string | `""` | 文件输出格式，取值同 `ConsoleFormat`。`""` 默认 `json`；非标准值视为掩码串 |
| `LogFilePath` | string | `""` | 日志文件路径，空串表示不输出到文件 |
| `MaxSize` | int | `100` | 单个日志文件最大大小（MB） |
| `MaxBackups` | int | `10` | 最多保留的日志备份数量 |
| `MaxAge` | int | `30` | 日志文件保留天数 |
| `Compress` | bool | `true` | 是否压缩备份文件 |

### 控制台格式字符串

`ConsoleFormat` 支持以下取值：

| 格式 | 说明 |
|------|------|
| `""` | 默认，回退到掩码格式 `"LCM"` |
| `"text"` | zap 内置 console 编码器：`<时间>\t<级别>\t<调用者>\t<消息>\t{...}`，人可读前缀以 tab 分隔，结构化字段以 JSON 对象追加在后。如 `2024-05-20T15:30:00.000+0800\tINFO\tmain.go:20\tuser logged in\t{"uid":42}` |
| `"json"` | zap JSON 编码器（slog json 风格）：`{"level":"info","ts":"...","caller":"main.go:20","msg":"..."}` |
| `"off"` | 关闭控制台输出 |
| 其它非空值 | 视为掩码串（旧版兼容），如 `"TLCM"`、`"LM"` |

掩码格式由以下字符组合而成（`ConsoleFormat` / `LogFileFormat` 设为非标准值时即作为掩码处理）：

| 字符 | 含义 | 示例输出 |
|------|------|----------|
| `T` | 时间 (ISO8601) | `2024-05-20T15:30:00.000+0800` |
| `L` | 日志级别 | `INFO` / `ERROR` |
| `C` | 调用者位置 | `main.go:20` |
| `M` | 日志消息 | 实际日志内容 |

示例：`"TLCM"` 输出完整信息，`"LM"` 仅输出级别和消息，`"off"` 关闭控制台。

### 配置函数

```go
zaplogs.NewLogConfig(level, logFile, consoleFormat string) LogConfig
zaplogs.NewLogConfigEmpty() LogConfig
```

`NewLogConfig` 是**旧版便捷构造函数**：第 3 个参数 `consoleFormat` 直接写入 `ConsoleFormat`（传入 `"LCM"` 即等价于 `ConsoleFormat: "LCM"`，作为掩码串处理），保持与早期版本的兼容。`NewLogConfig` 内部填充默认值：`MaxSize=100`、`MaxBackups=10`、`MaxAge=30`、`Compress=true`。`NewLogConfigEmpty` 返回全部默认值（`MaxBackups` 默认 `3`，其余与 `NewLogConfig` 一致）。

新代码建议直接使用 `LogConfig` 结构体字面量，或从 YAML 读取配置（见下节），以获得全部字段的控制权。

### YAML 配置

`LogConfig` 的每个字段都带有 `yaml` tag，可以直接通过标准 YAML 库反序列化。**库本身不提供 `LoadConfig` 函数** —— 由调用方自行读取并解析 YAML，再将 `LogConfig` 传入 `InitDefaultLogger` / `CreateLogger`。

示例 `config.yaml`：

```yaml
level: info            # debug / info / warn / error / fatal
console_format: text   # "" / text / json / off（"" 默认回退到掩码 "LCM"）；非标准值如 "TLCM" 视为掩码串
log_file_format: json  # "" 默认 json / text / json / off；非标准值视为掩码串
log_file_path: ./logs/app.log
max_size: 100          # 单文件最大 MB
max_backups: 10        # 最多保留备份数
max_age: 30            # 保留天数
compress: true         # 是否压缩备份文件
```

加载并初始化（需要调用方自行引入 YAML 库，如 `gopkg.in/yaml.v3`）：

```go
package main

import (
    "os"

    "gopkg.in/yaml.v3" // 调用方自行选择 YAML 库
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

## API 概览

### 全局函数（使用默认日志器）

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

> `DefaultLogger()` 返回默认的 `*Logger` 实例，可赋给自定义接口类型或用于依赖注入；`DefaultZapLogger()` 返回底层 `*zap.Logger`，供结构化日志调用方使用。

### 日志器管理

```go
zaplogs.CreateLogger(name string, config LogConfig) (*Logger, error)
zaplogs.GetLogger(name string) (*Logger, bool)
zaplogs.CloseAll() error
```

> `CloseAll()` 会先刷新所有日志器缓冲，再关闭底层日志文件（释放文件句柄），并重置默认日志器 —— 之后调用全局函数会自动重新初始化默认日志器。

### Logger 实例方法

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

> `ZapLogger()` / `SugaredLogger()` 返回底层 zap 原始对象，供需要结构化字段或 sugared API 的第三方库使用；`Close()` 关闭该日志器的底层文件，调用前应确保不再有并发日志写入。

## 示例

完整示例代码请查看 [examples](./examples) 目录：

- [basic](./examples/basic) — 零配置与自定义配置基础用法
- [multi-logger](./examples/multi-logger) — 多日志器管理
- [file-output](./examples/file-output) — 文件输出与日志轮转

## 依赖

- [go.uber.org/zap](https://github.com/uber-go/zap) — 高性能结构化日志
- [gopkg.in/natefinch/lumberjack.v2](https://github.com/natefinch/lumberjack) — 日志轮转

## License

MIT
