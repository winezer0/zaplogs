package zaplogs

// LogConfig holds logging configuration.
type LogConfig struct {
	// ConsoleLevel is the minimum log level for console output: debug, info, warn, error.
	// Empty string defaults to "info".
	ConsoleLevel string `yaml:"console_level"`
	// FileLevel is the minimum log level for file output: debug, info, warn, error.
	// Empty string defaults to "debug".
	FileLevel string `yaml:"file_level"`
	// ConsoleFormat is the console (stdout) output format.
	//   ""         - fallback to default mask "LCM"
	//   "text"     - console text format (zap console; structured fields appended as a JSON object)
	//   "json"     - json format
	//   "off"      - disable console output
	//   any other  - mask format, e.g. "TLCM", "LM" (T=time L=level C=caller M=message)
	ConsoleFormat string `yaml:"console_format"`

	// LogFileFormat is the file output format (same value options as ConsoleFormat).
	//   ""         - fallback to default "json"
	//   "text"     - console text format
	//   "json"     - json format
	//   "off"      - disable file output
	//   any other  - mask format
	LogFileFormat string `yaml:"log_file_format"`
	// LogFilePath is the log file path; empty means no file output.
	LogFilePath string `yaml:"log_file_path"`
	// MaxSize is the maximum size in megabytes of a single log file before rotation.
	MaxSize int `yaml:"max_size"`
	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int `yaml:"max_backups"`
	// MaxAge is the maximum number of days to retain old log files.
	MaxAge int `yaml:"max_age"`
	// Compress determines whether rotated files are compressed.
	Compress bool `yaml:"compress"`
}

// NewConfig 创建日志配置实例，提供默认值
func NewConfig(level, logFile, consoleFormat string) LogConfig {
	if level == "" {
		level = "info" // 默认info级别
	}
	if consoleFormat == "" {
		consoleFormat = "LCM"
	}
	return LogConfig{
		ConsoleLevel:  level,
		ConsoleFormat: consoleFormat,

		FileLevel:     "debug",
		LogFileFormat: "json",
		LogFilePath:   logFile,

		MaxSize:    100,  // 单个文件最大100MB
		MaxBackups: 3,    // 最多保留10个备份
		MaxAge:     30,   // 保留30天
		Compress:   true, // 压缩备份文件
	}
}

// DefaultConfig 创建日志配置实例，提供全部默认值
func DefaultConfig() LogConfig {
	return NewConfig("info", "", "LCM")
}
