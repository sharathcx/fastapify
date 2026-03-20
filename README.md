# Fastapify

![Fastapify Swagger UI](image.png)

Fastapify is a minimalist Go module built on top of [Gin](https://gin-gonic.com/) that provides automatic request validation/binding and OpenAPI (Swagger) documentation generation. It simplifies routing with a chainable builder API, auto-validation middleware, and a generic `Req[T]()` helper to retrieve bound request data.

## Installation

```bash
go get github.com/sharathcx/fastapify
```

## Setup & Example Usage

Here is a complete example of setting up a CRUD API for managing "Users" using Fastapify. You can find the runnable code in the [examples/user-api](examples/user-api) directory.

### 1. Define your Models

```go
package controllers

type User struct {
	ID    int    `json:"id" uri:"id"`
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type CreateUserReq struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type UserIdReq struct {
	ID int `uri:"id" binding:"required"`
}

// Use the "form" tag for query parameters
type ListUsersQuery struct {
	Search string `form:"search"`
	Limit  int    `form:"limit"`
}

// Combine URI params and JSON body params in a single struct
type UpdateReqCombined struct {
	ID    int    `uri:"id" binding:"required"` // from URI
	Name  string `json:"name"`                 // from JSON body
	Email string `json:"email"`                // from JSON body
}
```

### 2. Create the Controller Handlers

Handlers are standard `gin.HandlerFunc` functions. Use `fastapify.Req[T](c)` to retrieve the automatically validated and bound request data.

```go
package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/sharathcx/fastapify"
)

var users = []User{
	{ID: 1, Name: "John Doe", Email: "john@example.com"},
}
var nextID = 2

// GetUser - GET /users/{id}
func GetUser(c *gin.Context) {
	req := fastapify.Req[UserIdReq](c)

	for _, u := range users {
		if u.ID == req.ID {
			c.JSON(200, u)
			return
		}
	}
	c.JSON(404, gin.H{"error": "User not found"})
}

// CreateUser - POST /users
func CreateUser(c *gin.Context) {
	req := fastapify.Req[CreateUserReq](c)

	newUser := User{
		ID:    nextID,
		Name:  req.Name,
		Email: req.Email,
	}
	users = append(users, newUser)
	nextID++
	c.JSON(200, newUser)
}

// UpdateUser - PATCH /users/{id}
func UpdateUser(c *gin.Context) {
	req := fastapify.Req[UpdateReqCombined](c)

	for i, u := range users {
		if u.ID == req.ID {
			if req.Name != "" {
				users[i].Name = req.Name
			}
			if req.Email != "" {
				users[i].Email = req.Email
			}
			c.JSON(200, users[i])
			return
		}
	}
	c.JSON(404, gin.H{"error": "User not found"})
}

// DeleteUser - DELETE /users/{id}
func DeleteUser(c *gin.Context) {
	req := fastapify.Req[UserIdReq](c)

	for i, u := range users {
		if u.ID == req.ID {
			users = append(users[:i], users[i+1:]...)
			c.JSON(200, gin.H{"message": "Deleted"})
			return
		}
	}
	c.JSON(404, gin.H{"error": "User not found"})
}
```

### 3. Register Routes in Main

Wrap your Gin router with `fastapify.New()` and use the chainable builder to register routes with `.Body()` and `.Response()` for OpenAPI schema generation.

```go
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sharathcx/fastapify"
	"yourproject/controllers"
)

func main() {
	r := gin.Default()
	app := fastapify.New(r)

	// Register Routes with schema declarations for Swagger
	app.GET("/users/{id}", controllers.GetUser).
		Body(controllers.UserIdReq{}).
		Response(controllers.User{})

	app.POST("/users", controllers.CreateUser).
		Body(controllers.CreateUserReq{}).
		Response(controllers.User{})

	app.PATCH("/users/{id}", controllers.UpdateUser).
		Body(controllers.UpdateReqCombined{}).
		Response(controllers.User{})

	app.DELETE("/users/{id}", controllers.DeleteUser).
		Body(controllers.UserIdReq{})

	// Setup Swagger UI
	// This will generate docs at /openapi.json and serve the UI at /docs
	app.SetupSwagger("/openapi.json")

	log.Println("Server running on http://localhost:8080")
	log.Println("Swagger docs available at http://localhost:8080/docs")
	r.Run(":8080")
}
```

## How It Works

### Auto-Validation Middleware

When you register a route with `.Body(MyStruct{})`, Fastapify automatically injects validation middleware that:

1. Binds URI parameters first and snapshots their values
2. Binds the request body (POST/PUT/PATCH) or query params (GET/DELETE)
3. Restores URI values to prevent body payloads from overriding path parameters
4. Returns a `422` validation error if binding fails
5. Stores the validated struct in the Gin context

Your handler then retrieves the pre-validated data with `fastapify.Req[T](c)` — no manual binding needed.

### Manual Binding

If you prefer to handle binding yourself, you can use `fastapify.Bind()` directly:

```go
func MyHandler(c *gin.Context) {
	var req MyRequest
	if !fastapify.Bind(c, &req) {
		return // error response already sent
	}
	// use req...
}
```

## Error Handling

You can return structured HTTP errors using `fastapify.NewApiError`:

```go
fastapify.NewApiError(404, "Item not found", fastapify.ErrNotFound, nil)
```

Available error codes: `ErrValidation`, `ErrBadRequest`, `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, `ErrResourceConflict`, `ErrUploadError`, `ErrInternalError`.

## Features

- **Auto-Validation Middleware:** Requests are automatically validated and bound before your handler runs. Use `fastapify.Req[T](c)` to access the result.
- **URI Parameter Protection:** URI parameters are bound first and protected from being overridden by the request body.
- **Chainable Route Builder:** `app.GET("/path", handler).Body(ReqStruct{}).Response(RespStruct{})` for clean route registration with OpenAPI schema.
- **Auto Swagger Generation:** Inspects your structs and automatically builds an OpenAPI 3.0 specification with Swagger UI.
- **Flexible Route Syntax:** Supports both Gin-style `:id` and OpenAPI-style `{id}` parameters, normalized automatically.
- **Standardized Error Handling:** Consistent `ApiError` structure with built-in validation error formatting.
- **Full HTTP Support:** GET, POST, PUT, PATCH, and DELETE methods.
- **Query Parameter Binding:** Use the `form` tag in your structs to bind query string parameters.
- **Middleware Support:** Pass per-route middleware: `app.GET("/path", handler, authMiddleware)`.
- **Timeout Middleware:** Built-in `fastapify.TimeoutMiddleware(duration)` for request timeouts.
