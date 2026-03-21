package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/sharathcx/fastapify"
)

// User model
type User struct {
	ID    int    `json:"id" uri:"id"`
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

// Request models
type CreateUserReq struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type UserIdReq struct {
	ID int `uri:"id" binding:"required"`
}

type ListUsersQuery struct {
	Search string `form:"search"`
	Limit  int    `form:"limit"`
}

type UpdateReqCombined struct {
	ID    int    `uri:"id" binding:"required"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users = []User{
	{ID: 1, Name: "John Doe", Email: "john@example.com"},
}
var nextID = 2

func sendError(c *gin.Context, statusCode int, message, code string) {
	err := fastapify.NewApiError(statusCode, message, code, nil)
	status, res := fastapify.HandleError(err)
	c.JSON(status, res)
}

// GetUser - GET /users/:id
func GetUser(c *gin.Context) any {
	req := fastapify.Req[UserIdReq](c)

	for _, u := range users {
		if u.ID == req.ID {
			return fastapify.NewApiResponse(200, u, "Success")
		}
	}
	return fastapify.NotFound("User not found")
}

// CreateUser - POST /users
func CreateUser(c *gin.Context) any {
	req := fastapify.Req[CreateUserReq](c)

	newUser := User{
		ID:    nextID,
		Name:  req.Name,
		Email: req.Email,
	}
	users = append(users, newUser)
	nextID++
	return fastapify.NewApiResponse(200, newUser, "Success")
}

// UpdateUser - PATCH /users/:id
func UpdateUser(c *gin.Context) any {
	req := fastapify.Req[UpdateReqCombined](c)

	for i, u := range users {
		if u.ID == req.ID {
			if req.Name != "" {
				users[i].Name = req.Name
			}
			if req.Email != "" {
				users[i].Email = req.Email
			}
			return fastapify.NewApiResponse(200, users[i], "Success")
		}
	}
	return fastapify.NotFound("User not found")
}

// DeleteUser - DELETE /users/:id
func DeleteUser(c *gin.Context) any {
	req := fastapify.Req[UserIdReq](c)

	for i, u := range users {
		if u.ID == req.ID {
			users = append(users[:i], users[i+1:]...)
			return fastapify.NewApiResponse[*User](200, nil, "Success")
		}
	}
	return fastapify.NotFound("User not found")
}
