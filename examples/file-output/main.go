package main

import (
	"fmt"
	"os"
	"path/filepath"

	"zaplogs"
)

func main() {
	fmt.Println("=== 文件输出与日志轮转示例 ===")

	// 日志输出目录
	logDir := "./logs"

	// ==================== 仅文件输出（关闭控制台） ====================
	fmt.Println("\n--- 仅文件输出 ---")

	// 控制台格式设为 "off" 关闭控制台输出，日志仅写入文件
	fileOnlyConfig := zaplogs.NewLogConfig("debug", filepath.Join(logDir, "app.log"), "off")
	fileLogger, err := zaplogs.CreateLogger("file-only", fileOnlyConfig)
	if err != nil {
		panic(fmt.Sprintf("创建文件日志器失败: %v", err))
	}

	fileLogger.Debugf("这条日志只写入文件，不在控制台显示")
	fileLogger.Infof("应用启动完成")
	fileLogger.Warnf("配置文件使用默认值: %s", "config.yaml")
	_ = fileLogger.Sync()

	fmt.Printf("日志已写入: %s\n", filepath.Join(logDir, "app.log"))

	// ==================== 同时输出到控制台和文件 ====================
	fmt.Println("\n--- 控制台 + 文件双输出 ---")

	dualConfig := zaplogs.NewLogConfig("info", filepath.Join(logDir, "dual.log"), "TLCM")
	dualLogger, err := zaplogs.CreateLogger("dual-output", dualConfig)
	if err != nil {
		panic(fmt.Sprintf("创建双输出日志器失败: %v", err))
	}

	dualLogger.Infof("这条日志同时输出到控制台和文件")
	dualLogger.Errorf("模拟错误: 数据库连接超时")
	_ = dualLogger.Sync()

	// ==================== 自定义轮转参数 ====================
	fmt.Println("\n--- 自定义轮转参数 ---")

	// 手动配置轮转参数
	customConfig := zaplogs.LogConfig{
		Level:         "debug",
		LogFile:       filepath.Join(logDir, "custom.log"),
		ConsoleFormat: "LM",   // 控制台仅显示级别+消息
		MaxSize:       1,      // 单个文件最大 1MB（便于测试轮转）
		MaxBackups:    5,      // 最多保留 5 个备份
		MaxAge:        7,      // 保留 7 天
		Compress:      false,  // 不压缩（便于查看）
	}

	customLogger, err := zaplogs.CreateLogger("custom-rotation", customConfig)
	if err != nil {
		panic(fmt.Sprintf("创建自定义轮转日志器失败: %v", err))
	}

	// 写入多条日志
	for i := 1; i <= 10; i++ {
		customLogger.Infof("轮转测试日志 #%d: 当文件超过 MaxSize 时自动切割", i)
	}
	_ = customLogger.Sync()

	fmt.Printf("自定义轮转日志: %s\n", filepath.Join(logDir, "custom.log"))

	// ==================== 验证文件生成 ====================
	fmt.Println("\n--- 生成的日志文件 ---")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		fmt.Printf("读取日志目录失败: %v\n", err)
	} else {
		for _, entry := range entries {
			info, _ := entry.Info()
			fmt.Printf("  %s (%d bytes)\n", entry.Name(), info.Size())
		}
	}

	// 关闭所有日志器
	_ = zaplogs.CloseAll()
	fmt.Println("\n所有日志器已关闭，示例结束。")
}
