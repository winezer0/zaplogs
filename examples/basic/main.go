package main

import (
	"fmt"

	"github.com/winezer0/zaplogs"
)

func main() {
	// ==================== 零配置使用 ====================
	// 无需任何初始化，直接使用全局函数
	fmt.Println("=== 零配置使用 ===")
	zaplogs.Debug("这是一条调试日志")
	zaplogs.Info("这是一条信息日志")
	zaplogs.Warn("这是一条警告日志")
	zaplogs.Errorf("这是一条错误日志: %s", "something went wrong")

	// 格式化输出
	zaplogs.Infof("服务启动成功，端口: %d, 环境: %s", 8080, "production")

	// ==================== 自定义配置 ====================
	fmt.Println("\n=== 自定义配置 ===")

	// 创建配置: 级别=debug, 无文件输出, 控制台格式=TLCM(时间+级别+调用者+消息)
	config := zaplogs.NewConfig("debug", "", "TLCM")

	// 初始化默认日志器（覆盖自动创建的默认配置）
	if err := zaplogs.InitDefaultLogger(config); err != nil {
		panic(err)
	}

	// 现在 debug 级别的日志也会输出
	zaplogs.Debugf("调试信息: 当前用户数=%d", 42)
	zaplogs.Infof("用户 %s 登录成功", "admin")
	zaplogs.Warnf("内存使用率: %.1f%%", 85.5)

	// 刷新缓冲区
	_ = zaplogs.Sync()
}
