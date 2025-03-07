package main

import (
	"myappg/config"
	"myappg/routers"
	"myappg/utils"

	"go.uber.org/zap"
)

func main() {
	// Initialize configuration
	config.InitConfig()

	// Initialize logger
	utils.InitLogger()
	// utils.InitLoggerFile()
	// utils.InitAsyncLoggerFile()

	defer utils.Logger.Sync() // Ensure the log buffer is flushed

	// Initialize database
	utils.InitMongoDB()
	// defer utils.mongoClient.Disconnect(context.Background())

	// Set up routes
	r := routers.SetupRouter()

	// Start server and listen for incoming requests
	utils.Logger.Info("Starting server on port " + config.AppConfig.Server.Port)
	if err := r.Run(":" + config.AppConfig.Server.Port); err != nil {
		utils.Logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// main function is the entry point of the application
