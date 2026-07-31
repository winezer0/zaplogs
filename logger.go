package zaplogs

import (
	"fmt"
	"io"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// -------------------------- 日志器实现 --------------------------

// Logger 日志器实例，线程安全
type Logger struct {
	zapLogger *zap.Logger
	config    LogConfig
	mu        sync.RWMutex
	closer    io.Closer // 底层文件(如 lumberjack)，关闭时释放文件句柄
}

// init 初始化日志器核心(已修正EncodeTime配置)
func (l *Logger) init() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 解析日志级别
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(l.config.Level)); err != nil {
		level = zapcore.InfoLevel // 默认为info级别
	}

	// 准备输出核心
	var cores []zapcore.Core

	// 控制台输出
	if l.config.ConsoleFormat != "" && l.config.ConsoleFormat != "off" {
		encoder := newConsoleEncoder(l.config.ConsoleFormat)
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.Lock(os.Stdout),
			level,
		))
	}

	// 文件输出(带日志轮转，正确配置时间格式)
	if l.config.LogFile != "" {
		if err := ensureDir(l.config.LogFile); err != nil {
			return fmt.Errorf("failed to create log dir: %w", err)
		}

		// 日志轮转配置
		rotator := &lumberjack.Logger{
			Filename:   l.config.LogFile,
			MaxSize:    l.config.MaxSize,    // 单个文件最大100MB
			MaxBackups: l.config.MaxBackups, // 最多保留10个备份
			MaxAge:     l.config.MaxAge,     // 保留30天
			Compress:   l.config.Compress,   // 压缩备份文件
		}

		// 配置文件日志编码器(含时间格式)
		fileEncoderCfg := zap.NewProductionEncoderConfig()
		// 配置时间格式为ISO8601(如：2024-05-20T15:30:00.000Z)
		fileEncoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		// 创建JSON格式编码器
		fileEncoder := zapcore.NewJSONEncoder(fileEncoderCfg)
		cores = append(cores, zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(rotator),
			level,
		))
		l.closer = rotator
	}

	if len(cores) == 0 {
		return fmt.Errorf("no log output (console/file) has been configured")
	}

	// 创建zap日志器
	l.zapLogger = zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),      // 显示调用位置(如 main.go:20)
		zap.AddCallerSkip(2), // 跳过内部方法，显示真实业务代码位置
	)

	return nil
}

// 日志输出方法
// 注意：zap.Logger / zap.SugaredLogger 本身已是 goroutine-safe，无需额外锁
func (l *Logger) Debugf(template string, args ...interface{}) {
	if l.zapLogger != nil {
		l.zapLogger.Sugar().Debugf(template, args...)
	}
}

func (l *Logger) Infof(template string, args ...interface{}) {
	if l.zapLogger != nil {
		l.zapLogger.Sugar().Infof(template, args...)
	}
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	if l.zapLogger != nil {
		l.zapLogger.Sugar().Warnf(template, args...)
	}
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	if l.zapLogger != nil {
		l.zapLogger.Sugar().Errorf(template, args...)
	}
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	if l.zapLogger != nil {
		l.zapLogger.Sugar().Fatalf(template, args...)
	}
}

func (l *Logger) Debug(args ...interface{}) {
	if l.zapLogger != nil {
		l.zapLogger.Sugar().Debug(args...)
	}
}

func (l *Logger) Info(args ...interface{}) {
	if l.zapLogger != nil {
		l.zapLogger.Sugar().Info(args...)
	}
}

func (l *Logger) Warn(args ...interface{}) {
	if l.zapLogger != nil {
		l.zapLogger.Sugar().Warn(args...)
	}
}

func (l *Logger) Error(args ...interface{}) {
	if l.zapLogger != nil {
		l.zapLogger.Sugar().Error(args...)
	}
}

func (l *Logger) Fatal(args ...interface{}) {
	if l.zapLogger != nil {
		l.zapLogger.Sugar().Fatal(args...)
	}
}

// Sync 刷新日志缓冲区
func (l *Logger) Sync() error {
	if l.zapLogger != nil {
		return l.zapLogger.Sync()
	}
	return nil
}

// Close 关闭日志器底层资源（如日志文件），释放文件句柄
// 调用前应确保不再有并发日志写入
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closer != nil {
		err := l.closer.Close()
		l.closer = nil
		return err
	}
	return nil
}

// ZapLogger returns the underlying zap logger for libraries that use structured zap fields.
func (l *Logger) ZapLogger() *zap.Logger {
	if l.zapLogger == nil {
		return zap.NewNop()
	}
	return l.zapLogger
}

// SugaredLogger returns the underlying zap.SugaredLogger for libraries that use the sugared API.
func (l *Logger) SugaredLogger() *zap.SugaredLogger {
	if l.zapLogger == nil {
		return zap.NewNop().Sugar()
	}
	return l.zapLogger.Sugar()
}