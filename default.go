package zaplogs

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

var defaultLogger *Logger       // 旧版本默认日志器
var defaultLoggerOnce sync.Once // 用于确保 defaultLogger 只初始化一次

// InitDefaultLogger 旧版本初始化函数，兼容老代码
func InitDefaultLogger(config LogConfig) error {
	var err error
	defaultLogger, err = CreateLogger("default", config)
	return err
}

// NewDefaultLogger 旧版本初始化函数，兼容老代码
func NewDefaultLogger(level, logFile, consoleFormat string) error {
	return InitDefaultLogger(NewLogConfig(level, logFile, consoleFormat))
}

// ensureDefaultLogger 确保 defaultLogger 已初始化，失败时用 Nop 兜底，保证永不为 nil
func ensureDefaultLogger() {
	if defaultLogger == nil {
		defaultLoggerOnce.Do(func() {
			if err := InitDefaultLogger(NewLogConfigEmpty()); err != nil {
				fmt.Printf("init logger error: %v\n", err)
				// fallback: Nop logger 兜底，保证 defaultLogger 永不为 nil
				defaultLogger = &Logger{
					zapLogger: zap.NewNop(),
				}
			}
		})
	}
}

// Sync 旧版本全局刷新函数
func Sync() error {
	ensureDefaultLogger()
	return defaultLogger.Sync()
}

// DefaultLogger returns the default *Logger instance for use with interface types.
func DefaultLogger() *Logger {
	ensureDefaultLogger()
	return defaultLogger
}

// DefaultZapLogger returns the configured default logger for structured zap consumers.
func DefaultZapLogger() *zap.Logger {
	ensureDefaultLogger()
	return defaultLogger.ZapLogger()
}

// 旧版本全局日志函数，直接转发到 default 日志器
func Debugf(template string, args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Fatalf(template, args...)
}

func Debug(args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Debug(args...)
}

func Info(args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Info(args...)
}

func Warn(args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Warn(args...)
}

func Error(args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Error(args...)
}

func Fatal(args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Fatal(args...)
}