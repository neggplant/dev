package services

import (
	"context"
	"errors"
	"myappg/config"
	"myappg/models"
	"myappg/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type UserService struct {
	collection *mongo.Collection
}

func NewUserService() *UserService {
	db := utils.GetMongoDB(config.AppConfig.MongoDB.UserDB)
	return &UserService{
		collection: db.Collection("users"),
	}
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	var users []models.User

	cursor, err := s.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) CreateUser(user *models.User) error {
	_, err := s.collection.InsertOne(context.Background(), user)
	return err
}

func (s *UserService) GetUserByID(id string) (*models.User, error) {
	var user models.User
	cacheKey := "user:" + id

	// 使用通用查询方法
	err := utils.CacheFirstQuery(
		context.Background(),
		config.AppConfig.Redis.UserDB, // Redis 用户库
		cacheKey,
		config.AppConfig.MongoDB.UserDB, // MongoDB 用户库
		"users",
		&user,
		func(collection *mongo.Collection) error {
			objectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				return err
			}
			return collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
		},
		1*time.Hour, // 缓存过期时间
	)

	return &user, err
}

var (
	ErrUserNotFound = errors.New("user not found")
)

func (s *UserService) UpdateUser(id string, updateData *models.User) (*models.User, error) {
	// 转换ID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err // 由控制器处理
	}

	// 执行更新
	result, err := s.collection.UpdateByID(
		context.Background(),
		objectID,
		bson.M{"$set": updateData},
	)
	if err != nil {
		return nil, err
	}

	// 检查是否实际更新了文档
	if result.MatchedCount == 0 {
		return nil, ErrUserNotFound
	}

	// 使缓存失效
	cacheKey := "user:" + id
	if err := utils.GetRedisClient(config.AppConfig.Redis.UserDB).
		Del(context.Background(), cacheKey).Err(); err != nil {
		utils.Logger.Warn("Cache invalidation failed",
			zap.String("key", cacheKey),
			zap.Error(err))
	}

	// 返回更新后的完整数据
	return s.GetUserByID(id)
}

func (s *UserService) DeleteUser(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = s.collection.DeleteOne(context.Background(), bson.M{"_id": objectID})

	cacheKey := "user:" + id
	if err := utils.GetRedisClient(config.AppConfig.Redis.UserDB).
		Del(context.Background(), cacheKey).Err(); err != nil {
		utils.Logger.Warn("Cache invalidation failed",
			zap.String("key", cacheKey),
			zap.Error(err))
	}
	return err
}
