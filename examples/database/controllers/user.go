package controllers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/sharathcx/fastapify"
	"gorm.io/gorm"
)

// UserController holds the DB connection.
type UserController struct {
	db *gorm.DB
}

// NewUserController initializes the database connection and runs migrations.
func NewUserController(db *gorm.DB) *UserController {
	// Auto Migrate the User model
	db.AutoMigrate(&User{})

	return &UserController{db: db}
}

type User struct {
	ID    uint   `json:"id" gorm:"primaryKey" uri:"id"`
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required" gorm:"unique"`
}


type CreateUserReq struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type UpdateUserReq struct {
	ID    uint   `uri:"id" binding:"required"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserIdReq struct {
	ID uint `uri:"id" binding:"required"`
}

// Empty Request for listing users (no body/uri required)
type ListUsersReq struct{}



func (c *UserController) GetUser(ctx *gin.Context, req *UserIdReq) (*User, error) {
	var user User
	if err := c.db.First(&user, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fastapify.NewApiError(404, "User not found", fastapify.ErrNotFound, nil)
		}
		return nil, fastapify.NewApiError(500, "Internal Server Error", fastapify.ErrInternalError, []any{err.Error()})
	}
	return &user, nil
}

func (c *UserController) CreateUser(ctx *gin.Context, req *CreateUserReq) (*User, error) {
	newUser := User{
		Name:  req.Name,
		Email: req.Email,
	}

	if err := c.db.Create(&newUser).Error; err != nil {
		return nil, fastapify.NewApiError(400, "Could not create user, email might already exist", fastapify.ErrBadRequest, []any{err.Error()})
	}

	return &newUser, nil
}

func (c *UserController) UpdateUser(ctx *gin.Context, req *UpdateUserReq) (*User, error) {
	var user User
	if err := c.db.First(&user, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fastapify.NewApiError(404, "User not found", fastapify.ErrNotFound, nil)
		}
		return nil, fastapify.NewApiError(500, "Internal Server Error", fastapify.ErrInternalError, []any{err.Error()})
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := c.db.Save(&user).Error; err != nil {
		return nil, fastapify.NewApiError(500, "Could not save user", fastapify.ErrInternalError, []any{err.Error()})
	}

	return &user, nil
}

func (c *UserController) DeleteUser(ctx *gin.Context, req *UserIdReq) (*struct{}, error) {
	result := c.db.Delete(&User{}, req.ID)
	if result.Error != nil {
		return nil, fastapify.NewApiError(500, "Failed to delete user", fastapify.ErrInternalError, []any{result.Error.Error()})
	}

	if result.RowsAffected == 0 {
		return nil, fastapify.NewApiError(404, "User not found", fastapify.ErrNotFound, nil)
	}

	return nil, nil // Return empty response for successful deletion
}

func (c *UserController) ListUsers(ctx *gin.Context, req *ListUsersReq) (*[]User, error) {
	var users []User
	if err := c.db.Find(&users).Error; err != nil {
		return nil, fastapify.NewApiError(500, "Failed to fetch users", fastapify.ErrInternalError, []any{err.Error()})
	}
	return &users, nil
}