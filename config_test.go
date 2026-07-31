package zaplogs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

// TestNewLogConfig 测试日志配置创建
func TestNewLogConfig(t *testing.T) {
	config := NewConfig("debug", "test.log", "TLCM")
	if config.ConsoleLevel != "debug" {
		t.Errorf("Expected console level 'debug', got '%s'", config.ConsoleLevel)
	}
	if config.LogFilePath != "test.log" {
		t.Errorf("Expected log file path 'test.log', got '%s'", config.LogFilePath)
	}
	if config.ConsoleFormat != "TLCM" {
		t.Errorf("Expected console format 'TLCM', got '%s'", config.ConsoleFormat)
	}
	if config.MaxSize != 100 {
		t.Errorf("Expected max size 100, got %d", config.MaxSize)
	}
	if config.MaxBackups != 3 {
		t.Errorf("Expected max backups 3, got %d", config.MaxBackups)
	}
	if config.MaxAge != 30 {
		t.Errorf("Expected max age 30, got %d", config.MaxAge)
	}
	if !config.Compress {
		t.Error("Expected compress to be true")
	}
}

// TestNewLogConfigEmpty 测试空配置创建
func TestNewLogConfigEmpty(t *testing.T) {
	config := DefaultConfig()
	if config.ConsoleLevel != "info" {
		t.Errorf("Expected console level 'info', got '%s'", config.ConsoleLevel)
	}
	if config.LogFilePath != "" {
		t.Errorf("Expected empty log file path, got '%s'", config.LogFilePath)
	}
	if config.ConsoleFormat != "LCM" {
		t.Errorf("Expected console format 'LCM', got '%s'", config.ConsoleFormat)
	}
	if config.MaxSize != 100 {
		t.Errorf("Expected max size 100, got %d", config.MaxSize)
	}
	if config.MaxBackups != 3 {
		t.Errorf("Expected max backups 3, got %d", config.MaxBackups)
	}
	if config.MaxAge != 30 {
		t.Errorf("Expected max age 30, got %d", config.MaxAge)
	}
	if !config.Compress {
		t.Error("Expected compress to be true")
	}
}

// TestResolveOutput 测试单字段输出格式解析逻辑
func TestResolveOutput(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		defKind  string
		defMask  string
		wantKind string
		wantMask string
	}{
		{"console default (empty)", "", "mask", "LCM", "mask", "LCM"},
		{"text format", "text", "mask", "LCM", "text", ""},
		{"json format", "json", "mask", "LCM", "json", ""},
		{"off format", "off", "mask", "LCM", "off", ""},
		{"uppercase text", "TEXT", "mask", "LCM", "text", ""},
		{"uppercase json", "JSON", "mask", "LCM", "json", ""},
		{"uppercase off", "OFF", "mask", "LCM", "off", ""},
		{"mask string TLCM", "TLCM", "mask", "LCM", "mask", "TLCM"},
		{"mask string LM", "LM", "mask", "LCM", "mask", "LM"},
		{"lowercase mask string", "tlcm", "mask", "LCM", "mask", "TLCM"},
		{"file default json", "", "json", "", "json", ""},
		{"file text", "text", "json", "", "text", ""},
		{"file mask string", "TLCM", "json", "", "mask", "TLCM"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, maskValue := resolveOutput(tt.format, tt.defKind, tt.defMask)
			if kind != tt.wantKind {
				t.Errorf("resolveOutput() kind = %q, want %q", kind, tt.wantKind)
			}
			if maskValue != tt.wantMask {
				t.Errorf("resolveOutput() maskValue = %q, want %q", maskValue, tt.wantMask)
			}
		})
	}
}

// TestLoggerFormatVariants 测试不同控制台格式的日志器初始化
func TestLoggerFormatVariants(t *testing.T) {
	formats := []string{"json", "text"}
	for _, fmt := range formats {
		t.Run("console_"+fmt, func(t *testing.T) {
			config := LogConfig{
				ConsoleLevel:  "debug",
				ConsoleFormat: fmt,
				MaxSize:       100,
				MaxBackups:    3,
				MaxAge:        30,
				Compress:      true,
			}
			logger, err := CreateLogger("fmt-"+fmt, config)
			if err != nil {
				t.Fatalf("Failed to create logger with format %q: %v", fmt, err)
			}
			logger.Infof("test message with %s format", fmt)
			_ = logger.Sync()
		})
	}
	defer func() {
		_ = CloseAll()
	}()
}

// TestTextEncoder 测试文本编码器输出（zap console 格式，含消息内容）
func TestTextEncoder(t *testing.T) {
	enc := newTextEncoder()
	ent := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Date(2024, 5, 20, 15, 30, 0, 0, time.UTC),
		Message: "hello world",
	}
	buf, err := enc.EncodeEntry(ent, nil)
	if err != nil {
		t.Fatalf("EncodeEntry failed: %v", err)
	}
	defer buf.Free()

	output := buf.String()
	if !strings.Contains(output, "hello world") {
		t.Errorf("expected output to contain message 'hello world', got: %s", output)
	}
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected output to contain 'INFO', got: %s", output)
	}
}

// TestFileOutputWithFormats 测试文件输出配合不同格式
func TestFileOutputWithFormats(t *testing.T) {
	tmpDir := os.TempDir()
	logFile := filepath.Join(tmpDir, "test_format_output.log")
	defer os.Remove(logFile)

	config := LogConfig{
		ConsoleLevel:  "debug",
		ConsoleFormat: "off",
		LogFilePath:   logFile,
		LogFileFormat: "json",
		MaxSize:       100,
		MaxBackups:    3,
		MaxAge:        30,
		Compress:      true,
	}
	logger, err := CreateLogger("filefmt", config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.Infof("file format test message")
	_ = logger.Sync()
	defer func() {
		_ = CloseAll()
	}()

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}
