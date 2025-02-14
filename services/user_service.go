package services

import (
	"context"
	"myappg/config"
	"myappg/models"
	"myappg/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func (s *UserService) UpdateUser(id string, user *models.User) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"name":  user.Name,
			"email": user.Email,
		},
	}

	_, err = s.collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	return err
}

func (s *UserService) DeleteUser(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = s.collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	return err
}
