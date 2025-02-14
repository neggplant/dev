package utils

import (
	"context"
	"encoding/json"
	"myappg/config"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
)

var redisClients = make(map[int]*redis.Client) // 按数据库编号存储客户端

// 获取指定数据库的 Redis 客户端
func GetRedisClient(db int) *redis.Client {
	if client, ok := redisClients[db]; ok {
		return client
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.AppConfig.Redis.Host + ":" + config.AppConfig.Redis.Port,
		Password: config.AppConfig.Redis.Password,
		DB:       db,
	})

	// 测试连接
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic("Failed to connect to Redis DB " + strconv.Itoa(db) + ": " + err.Error())
	}

	redisClients[db] = client
	return client
}

// 通用查询封装：优先查 Redis，不存在则查 MongoDB，并回写 Redis
func CacheFirstQuery(
	ctx context.Context,
	redisDB int, // Redis 数据库编号
	cacheKey string, // Redis 缓存键
	mongoDBName string, // MongoDB 数据库名
	mongoCollection string, // MongoDB 集合名
	result interface{}, // 查询结果的指针
	mongoQuery func(*mongo.Collection) error, // MongoDB 查询逻辑
	expireTime time.Duration, // Redis 缓存过期时间
) error {
	// 1. 先查 Redis
	redisClient := GetRedisClient(redisDB)
	cachedData, err := redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		if err := json.Unmarshal(cachedData, result); err == nil {
			return nil // 缓存命中
		}
	}

	// 2. 缓存未命中，查 MongoDB
	mongoClient := GetMongoDB(mongoDBName)
	collection := mongoClient.Collection(mongoCollection)
	if err := mongoQuery(collection); err != nil {
		return err
	}

	// 3. 将结果写入 Redis
	dataToCache, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if err := redisClient.Set(ctx, cacheKey, dataToCache, expireTime).Err(); err != nil {
		return err
	}

	return nil
}
