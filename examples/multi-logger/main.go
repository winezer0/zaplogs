package main

import (
	"fmt"

	"github.com/winezer0/zaplogs"
)

func main() {
	fmt.Println("=== 多日志器管理示例 ===")

	// 创建业务日志器（仅控制台输出，格式: 级别+调用者+消息）
	bizConfig := zaplogs.NewLogConfig("debug", "", "LCM")
	bizLogger, err := zaplogs.CreateLogger("business", bizConfig)
	if err != nil {
		panic(fmt.Sprintf("创建业务日志器失败: %v", err))
	}

	// 创建访问日志器（仅控制台输出，格式: 时间+消息）
	accessConfig := zaplogs.NewLogConfig("info", "", "TM")
	accessLogger, err := zaplogs.CreateLogger("access", accessConfig)
	if err != nil {
		panic(fmt.Sprintf("创建访问日志器失败: %v", err))
	}

	// 使用业务日志器
	fmt.Println("\n--- 业务日志 ---")
	bizLogger.Debugf("开始处理订单: %s", "ORD-2024-001")
	bizLogger.Infof("订单创建成功: %s, 金额: %.2f", "ORD-2024-001", 199.99)
	bizLogger.Warnf("库存不足: SKU-%d, 剩余: %d", 10086, 3)

	// 使用访问日志器
	fmt.Println("\n--- 访问日志 ---")
	accessLogger.Infof("GET /api/orders 200 15ms")
	accessLogger.Infof("POST /api/orders 201 42ms")
	accessLogger.Infof("GET /api/users/1 404 3ms")

	// 通过名称获取已有日志器
	fmt.Println("\n--- 获取已有日志器 ---")
	logger, exists := zaplogs.GetLogger("business")
	if exists {
		logger.Info("通过 GetLogger 获取到业务日志器")
	}

	// 尝试获取不存在的日志器
	_, exists = zaplogs.GetLogger("nonexistent")
	fmt.Printf("日志器 'nonexistent' 是否存在: %v\n", exists)

	// 尝试创建同名日志器（会报错）
	_, err = zaplogs.CreateLogger("business", bizConfig)
	if err != nil {
		fmt.Printf("重复创建日志器报错: %v\n", err)
	}

	// 关闭所有日志器（刷新缓冲区并清空）
	fmt.Println("\n--- 关闭所有日志器 ---")
	if err := zaplogs.CloseAll(); err != nil {
		fmt.Printf("关闭日志器出错: %v\n", err)
	} else {
		fmt.Println("所有日志器已关闭")
	}
}
