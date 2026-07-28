# zaplogs

一个基于 [uber-go/zap](https://github.com/uber-go/zap) 的轻量级日志封装库，提供开箱即用的日志功能，支持控制台输出、文件轮转、多日志器管理。

[English](./Readme.EN.md)

## 特性

- **零配置启动** — 无需任何初始化即可直接使用全局日志函数
- **多日志器管理** — 按名称创建和管理多个独立日志器
- **日志轮转** — 基于 [lumberjack](https://github.com/natefinch/lumberjack) 实现自动日志切割与压缩
- **灵活的控制台格式** — 通过格式字符串自定义输出字段（时间/级别/调用者/消息）
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
    // 创建配置: 级别, 日志文件路径, 控制台格式
    config := zaplogs.NewLogConfig("debug", "./logs/app.log", "TLCM")
    
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
    // 创建业务日志器
    bizConfig := zaplogs.NewLogConfig("info", "./logs/biz.log", "LCM")
    bizLogger, err := zaplogs.CreateLogger("business", bizConfig)
    if err != nil {
        panic(err)
    }

    // 创建访问日志器
    accessConfig := zaplogs.NewLogConfig("info", "./logs/access.log", "TM")
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
| `LogFile` | string | `""` | 日志文件路径，空串表示不输出到文件 |
| `ConsoleFormat` | string | `"LCM"` | 控制台输出格式，`"off"` 或空串关闭控制台输出 |
| `MaxSize` | int | `100` | 单个日志文件最大大小（MB） |
| `MaxBackups` | int | `10` | 最多保留的日志备份数量 |
| `MaxAge` | int | `30` | 日志文件保留天数 |
| `Compress` | bool | `true` | 是否压缩备份文件 |

### 控制台格式字符串

格式字符串由以下字符组合而成：

| 字符 | 含义 | 示例输出 |
|------|------|----------|
| `T` | 时间 (ISO8601) | `2024-05-20T15:30:00.000+0800` |
| `L` | 日志级别 | `INFO` / `ERROR` |
| `C` | 调用者位置 | `main.go:20` |
| `M` | 日志消息 | 实际日志内容 |

示例：`"TLCM"` 输出完整信息，`"LM"` 仅输出级别和消息，`"off"` 关闭控制台。

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
```

### 日志器管理

```go
zaplogs.CreateLogger(name string, config LogConfig) (*Logger, error)
zaplogs.GetLogger(name string) (*Logger, bool)
zaplogs.CloseAll() error
```

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
```

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
