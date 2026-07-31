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

	// 解析控制台日志级别
	var consoleLevel zapcore.Level
	if l.config.ConsoleLevel != "" {
		if err := consoleLevel.UnmarshalText([]byte(l.config.ConsoleLevel)); err != nil {
			consoleLevel = zapcore.InfoLevel
		}
	} else {
		consoleLevel = zapcore.InfoLevel // 默认为info级别
	}

	// 解析文件日志级别
	var fileLevel zapcore.Level
	if l.config.FileLevel != "" {
		if err := fileLevel.UnmarshalText([]byte(l.config.FileLevel)); err != nil {
			fileLevel = zapcore.DebugLevel
		}
	} else {
		fileLevel = zapcore.DebugLevel // 默认为debug级别
	}

	// 准备输出核心
	var cores []zapcore.Core

	// 控制台输出
	consoleKind, consoleMask := resolveOutput(l.config.ConsoleFormat, "mask", "LCM")
	if consoleKind != "off" {
		var enc zapcore.Encoder
		switch consoleKind {
		case "json":
			enc = newJSONEncoder()
		case "text":
			enc = newTextEncoder()
		default: // "mask"
			enc = newMaskEncoder(consoleMask)
		}
		cores = append(cores, zapcore.NewCore(
			enc,
			zapcore.Lock(os.Stdout),
			consoleLevel,
		))
	}

	// 文件输出(带日志轮转，正确配置时间格式)
	if l.config.LogFilePath != "" {
		fileKind, fileMask := resolveOutput(l.config.LogFileFormat, "json", "")
		if fileKind != "off" {
			if err := ensureDir(l.config.LogFilePath); err != nil {
				return fmt.Errorf("failed to create log dir: %w", err)
			}

			// 日志轮转配置
			rotator := &lumberjack.Logger{
				Filename:   l.config.LogFilePath,
				MaxSize:    l.config.MaxSize,
				MaxBackups: l.config.MaxBackups,
				MaxAge:     l.config.MaxAge,
				Compress:   l.config.Compress,
			}

			// 文件编码器
			var fileEnc zapcore.Encoder
			switch fileKind {
			case "text":
				fileEnc = newTextEncoder()
			case "mask":
				fileEnc = newMaskEncoder(fileMask)
			default: // "json"
				fileEnc = newJSONEncoder()
			}
			cores = append(cores, zapcore.NewCore(
				fileEnc,
				zapcore.AddSync(rotator),
				fileLevel,
			))
			l.closer = rotator
		}
	}

	// 无任何输出目标（控制台与文件均关闭）时，使用 Nop logger 静默丢弃所有记录
	if len(cores) == 0 {
		l.zapLogger = zap.NewNop()
		return nil
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
