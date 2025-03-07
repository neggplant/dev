package controllers

import (
	"errors"
	"myappg/models"
	"myappg/services"
	"myappg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type UserController struct {
	userService *services.UserService
}

// NewUserController creates a new UserController instance
func NewUserController() *UserController {
	return &UserController{
		userService: services.NewUserService(),
	}
}

// GetUsers retrieves all users
func (ctrl *UserController) GetUsers(c *gin.Context) {
	users, err := ctrl.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (ctrl *UserController) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.userService.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (ctrl *UserController) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := ctrl.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (ctrl *UserController) UpdateUser(c *gin.Context) {
	// 1. Get user ID and validate format
	id := c.Param("id")
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
			"code":  "INVALID_ID",
		})
		return
	}

	// 2. Bind request data and validate
	var updateData models.User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		// More detailed field-level error information
		fieldErrors := make(map[string]string)
		for _, fieldErr := range err.(validator.ValidationErrors) {
			field := fieldErr.Field()
			switch fieldErr.Tag() {
			case "required":
				fieldErrors[field] = "This field is required"
			case "email":
				fieldErrors[field] = "Invalid email format"
			default:
				fieldErrors[field] = "Validation failed"
			}
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Validation failed",
			"fields": fieldErrors,
			"code":   "VALIDATION_ERROR",
		})
		return
	}

	// 3. Execute update operation
	updatedUser, err := ctrl.userService.UpdateUser(id, &updateData)
	if err != nil {
		// Return different status codes based on error type
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
				"code":  "USER_NOT_FOUND",
			})
		default:
			utils.Logger.Error("User update failed",
				zap.String("user_id", id),
				zap.Error(err))

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update user",
				"code":  "INTERNAL_ERROR",
			})
		}
		return
	}

	// 4. Return the complete updated data
	c.JSON(http.StatusOK, gin.H{
		"data": updatedUser,
		"meta": gin.H{
			"updated_at": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (ctrl *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.userService.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
