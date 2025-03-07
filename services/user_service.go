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

// UserService struct represents the user service
type UserService struct {
	collection *mongo.Collection
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	db := utils.GetMongoDB(config.AppConfig.MongoDB.UserDB)
	return &UserService{
		collection: db.Collection("users"),
	}
}

// GetAllUsers retrieves all users from the database
func (s *UserService) GetAllUsers() ([]models.User, error) {
	var users []models.User

	utils.Logger.Info("GetAllUsers called with")
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
	utils.Logger.Info("GetUserByID called with id:")

	// Use the general query method
	err := utils.CacheFirstQuery(
		context.Background(),
		config.AppConfig.Redis.UserDB, // Redis user database
		cacheKey,
		config.AppConfig.MongoDB.UserDB, // MongoDB user database
		"users",
		&user,
		func(collection *mongo.Collection) error {
			objectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				return err
			}
			return collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
		},
		1*time.Hour, // Cache expiration time
	)

	return &user, err
}

var (
	ErrUserNotFound = errors.New("user not found")
)

func (s *UserService) UpdateUser(id string, updateData *models.User) (*models.User, error) {
	// Convert ID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err // Handled by controller
	}

	// Execute update
	result, err := s.collection.UpdateByID(
		context.Background(),
		objectID,
		bson.M{"$set": updateData},
	)
	if err != nil {
		return nil, err
	}

	// Check if the document was actually updated
	if result.MatchedCount == 0 {
		return nil, ErrUserNotFound
	}

	// Invalidate cache
	cacheKey := "user:" + id
	if err := utils.GetRedisClient(config.AppConfig.Redis.UserDB).
		Del(context.Background(), cacheKey).Err(); err != nil {
		utils.Logger.Warn("Cache invalidation failed",
			zap.String("key", cacheKey),
			zap.Error(err))
	}

	// Return the complete updated data
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
