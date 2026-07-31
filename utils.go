package zaplogs

import (
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// -------------------------- 工具函数 --------------------------

// 创建掩码编码器（控制台格式 T/L/C/M）
func newMaskEncoder(format string) zapcore.Encoder {
	format = strings.ToUpper(format) // defensive: mask chars are case-insensitive
	cfg := zapcore.EncoderConfig{
		TimeKey:      "T",
		LevelKey:     "L",
		CallerKey:    "C",
		MessageKey:   "M",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	if !strings.Contains(format, "T") {
		cfg.TimeKey = ""
	}
	if !strings.Contains(format, "L") {
		cfg.LevelKey = ""
	}
	if !strings.Contains(format, "C") {
		cfg.CallerKey = ""
	}
	if !strings.Contains(format, "M") {
		cfg.MessageKey = ""
	}

	return zapcore.NewConsoleEncoder(cfg)
}

// newJSONEncoder 创建 JSON 编码器（ISO8601 时间格式）
func newJSONEncoder() zapcore.Encoder {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	return zapcore.NewJSONEncoder(cfg)
}

// resolveOutput resolves a single output format field to its effective mode.
// Returns one of: ("mask", maskValue), ("text",""), ("json",""), ("off","").
// defKind/defMask apply when format == "" (console default: "mask"/"LCM"; file default: "json"/"").
func resolveOutput(format, defKind, defMask string) (kind, maskValue string) {
	switch strings.ToLower(format) {
	case "off":
		return "off", ""
	case "json":
		return "json", ""
	case "text":
		return "text", ""
	case "":
		return defKind, defMask
	default:
		// Non-standard value treated as mask string; uppercase so that
		// "tlcm" == "TLCM" (matches slogs newMaskHandler normalization).
		return "mask", strings.ToUpper(format)
	}
}

// newTextEncoder 创建文本编码器（zap 内置 console 编码器，字段以 key=value 追加）
func newTextEncoder() zapcore.Encoder {
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	return zapcore.NewConsoleEncoder(cfg)
}

// ensureDir 确保目录存在
func ensureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}
