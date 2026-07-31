package zaplogs

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureStdout 临时将 os.Stdout 重定向到管道，返回恢复函数。
// 必须在创建 logger 之前调用（zap core 在 init 时锁定 os.Stdout 引用）。
func captureStdout(t *testing.T) func() string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("captureStdout: os.Pipe: %v", err)
	}
	os.Stdout = w
	return func() string {
		os.Stdout = old
		_ = w.Close()
		data, _ := io.ReadAll(r)
		_ = r.Close()
		return string(data)
	}
}

// TestDefaultConsoleFormatMask 空 ConsoleFormat 端到端回退到默认 mask "LCM"
func TestDefaultConsoleFormatMask(t *testing.T) {
	capture := captureStdout(t)
	config := LogConfig{ConsoleLevel: "debug", MaxSize: 100, MaxBackups: 3, MaxAge: 30, Compress: true}
	logger, err := CreateLogger("defconmask", config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer func() { _ = CloseAll() }()

	logger.Infof("default console mask message")
	out := capture()

	// mask "LCM"：级别 + 消息，无时间
	if !strings.Contains(out, "INFO") {
		t.Errorf("expected level INFO in output, got %q", out)
	}
	if !strings.Contains(out, "default console mask message") {
		t.Errorf("expected message in output, got %q", out)
	}
	if strings.HasPrefix(out, "2") {
		t.Errorf("expected no time prefix for mask LCM, got %q", out)
	}
}

// TestDefaultFileFormatJSON 空 LogFileFormat 端到端回退到默认 "json"
func TestDefaultFileFormatJSON(t *testing.T) {
	logFile := filepath.Join(os.TempDir(), "test_default_file_json.log")
	defer os.Remove(logFile)

	config := LogConfig{
		ConsoleLevel:  "debug",
		ConsoleFormat: "off",
		LogFilePath:   logFile,
		MaxSize:       100,
		MaxBackups:    3,
		MaxAge:        30,
		Compress:      true,
	}
	logger, err := CreateLogger("deffilejson", config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Infof("default file json message")
	_ = logger.Sync()
	_ = CloseAll() // 释放文件句柄后再读

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(line, "{") {
		t.Errorf("expected JSON line in file, got %q", line)
	}
	if !strings.Contains(line, "default file json message") {
		t.Errorf("expected message in file, got %q", line)
	}
}

// TestFormatCaseInsensitive mask 串与关键字大小写不敏感
func TestFormatCaseInsensitive(t *testing.T) {
	t.Run("lowercase mask", func(t *testing.T) {
		capture := captureStdout(t)
		config := LogConfig{ConsoleLevel: "debug", ConsoleFormat: "tlcm", MaxSize: 100, MaxBackups: 3, MaxAge: 30, Compress: true}
		logger, err := CreateLogger("ci-mask", config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer func() { _ = CloseAll() }()

		logger.Infof("lowercase mask message")
		out := capture()

		if !strings.Contains(out, "INFO") || !strings.Contains(out, "lowercase mask message") {
			t.Errorf("expected full mask output for lowercase 'tlcm', got %q", out)
		}
		if !strings.HasPrefix(out, "2") {
			t.Errorf("expected time prefix for mask TLCM, got %q", out)
		}
	})
	t.Run("uppercase keyword", func(t *testing.T) {
		capture := captureStdout(t)
		config := LogConfig{ConsoleLevel: "debug", ConsoleFormat: "TEXT", MaxSize: 100, MaxBackups: 3, MaxAge: 30, Compress: true}
		logger, err := CreateLogger("ci-text", config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer func() { _ = CloseAll() }()

		logger.Infof("uppercase text message")
		out := capture()

		if !strings.Contains(out, "uppercase text message") {
			t.Errorf("expected text output for uppercase 'TEXT', got %q", out)
		}
	})
}

// TestCloseAll 测试关闭所有日志器
func TestCloseAll(t *testing.T) {
	config := NewConfig("debug", "", "TLCM")
	_, err := CreateLogger("close1", config)
	if err != nil {
		t.Fatalf("Failed to create logger 1: %v", err)
	}

	_, err = CreateLogger("close2", config)
	if err != nil {
		t.Fatalf("Failed to create logger 2: %v", err)
	}

	err = CloseAll()
	if err != nil {
		t.Errorf("Failed to close all loggers: %v", err)
	}

	// 验证日志器已被清空
	_, exists := GetLogger("close1")
	if exists {
		t.Error("Expected logger to be closed")
	}
}

// TestLoggerWithFileOutput 测试带文件输出的日志器
func TestLoggerWithFileOutput(t *testing.T) {
	tmpDir := os.TempDir()
	logFile := filepath.Join(tmpDir, "test_file_output.log")
	defer os.Remove(logFile)

	config := NewConfig("debug", logFile, "off")
	logger, err := CreateLogger("filetest", config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		_ = CloseAll()
	}()

	logger.Infof("Test file output message")
	_ = logger.Sync()

	// 验证日志文件已创建
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}

// TestAllTargetsOff 所有输出目标关闭时静默丢弃，创建不报错（与 slogs 一致）
func TestAllTargetsOff(t *testing.T) {
	t.Run("console off no file", func(t *testing.T) {
		capture := captureStdout(t)
		config := LogConfig{ConsoleLevel: "debug", ConsoleFormat: "off", MaxSize: 100, MaxBackups: 3, MaxAge: 30, Compress: true}
		logger, err := CreateLogger("alloff-nofile", config)
		if err != nil {
			t.Fatalf("Expected nil error when all targets off, got: %v", err)
		}
		defer func() { _ = CloseAll() }()

		logger.Infof("should be silently discarded")
		out := capture()
		if out != "" {
			t.Errorf("expected no output when all targets off, got %q", out)
		}
	})
	t.Run("console and file both off", func(t *testing.T) {
		logFile := filepath.Join(os.TempDir(), "test_all_off.log")
		defer os.Remove(logFile)

		capture := captureStdout(t)
		config := LogConfig{
			ConsoleLevel:  "debug",
			ConsoleFormat: "off",
			LogFilePath:   logFile,
			LogFileFormat: "off",
			MaxSize:       100,
			MaxBackups:    3,
			MaxAge:        30,
			Compress:      true,
		}
		logger, err := CreateLogger("alloff-both", config)
		if err != nil {
			t.Fatalf("Expected nil error when all targets off, got: %v", err)
		}
		defer func() { _ = CloseAll() }()

		logger.Infof("should be silently discarded")
		out := capture()
		if out != "" {
			t.Errorf("expected no console output when all targets off, got %q", out)
		}
		if _, err := os.Stat(logFile); !os.IsNotExist(err) {
			t.Errorf("expected no log file to be created when file target off, stat err=%v", err)
		}
	})
}

// TestLoggerConcurrent 测试日志器并发安全性
func TestLoggerConcurrent(t *testing.T) {
	config := NewConfig("debug", "", "TLCM")
	logger, err := CreateLogger("concurrent", config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		_ = CloseAll()
	}()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Infof("Concurrent message %d", id)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
