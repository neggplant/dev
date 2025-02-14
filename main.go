package main

import (
	"myappg/config"
	"myappg/routers"
	"myappg/utils"

	"go.uber.org/zap"
)

func main() {
	// 初始化配置
	config.InitConfig()

	// 初始化日志
	utils.InitLogger()
	// utils.InitLoggerFile()

	defer utils.Logger.Sync() // 确保日志缓冲区刷新

	// 初始化数据库
	utils.InitMongoDB()
	// defer utils.mongoClient.Disconnect(context.Background())

	// 设置路由
	r := routers.SetupRouter()

	// 启动服务
	utils.Logger.Info("Starting server on port " + config.AppConfig.Server.Port)
	if err := r.Run(":" + config.AppConfig.Server.Port); err != nil {
		utils.Logger.Fatal("Failed to start server", zap.Error(err))
	}
}
