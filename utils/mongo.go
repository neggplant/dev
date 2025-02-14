package utils

import (
	"context"
	"myappg/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

// 初始化 MongoDB 连接
func InitMongoDB() {
	clientOptions := options.Client().ApplyURI(config.AppConfig.MongoDB.URI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic("Failed to connect to MongoDB: " + err.Error())
	}
	mongoClient = client
}

// 获取指定数据库的句柄
func GetMongoDB(dbName string) *mongo.Database {
	return mongoClient.Database(dbName)
}
