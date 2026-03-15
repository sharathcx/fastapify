package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/sharathcx/fastapify"
)

// Models
type User struct {
	ID    int    `json:"id" uri:"id"`
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type CreateUserReq struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type UpdateUserReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserIdReq struct {
	ID int `uri:"id" binding:"required"`
}

// In-memory DB for sample usercase
var users = []User{
	{ID: 1, Name: "John Cena", Email: "youcantseeme@example.com"},
}
var nextID = 2

func GetUser(c *gin.Context, req *UserIdReq) (*User, error) {
	for _, u := range users {
		if u.ID == req.ID {
			return &u, nil
		}
	}
	return nil, fastapify.NewApiError(404, "User not found", fastapify.ErrNotFound, nil)
}

func CreateUser(c *gin.Context, req *CreateUserReq) (*User, error) {
	newUser := User{
		ID:    nextID,
		Name:  req.Name,
		Email: req.Email,
	}
	users = append(users, newUser)
	nextID++
	return &newUser, nil
}

type UpdateReqCombined struct {
	ID    int    `uri:"id" binding:"required"` 
	Name  string `json:"name"`                 
	Email string `json:"email"`
}

func UpdateUser(c *gin.Context, req *UpdateReqCombined) (*User, error) {
	for i, u := range users {
		if u.ID == req.ID {
			if req.Name != "" {
				users[i].Name = req.Name
			}
			if req.Email != "" {
				users[i].Email = req.Email
			}
			return &users[i], nil
		}
	}
	return nil, fastapify.NewApiError(404, "User not found", fastapify.ErrNotFound, nil)
}

func DeleteUser(c *gin.Context, req *UserIdReq) (*struct{}, error) {
	for i, u := range users {
		if u.ID == req.ID {
			users = append(users[:i], users[i+1:]...)
			return nil, nil 
		}
	}
	return nil, fastapify.NewApiError(404, "User not found", fastapify.ErrNotFound, nil)
}
