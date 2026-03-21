# Fastapify

![Fastapify Swagger UI](image.png)

Fastapify is a minimalist Go module built on top of [Gin](https://gin-gonic.com/) that provides automatic request validation/binding, structured error handling, and OpenAPI (Swagger) documentation generation. It simplifies routing with a chainable builder API, auto-validation middleware, and a custom `HandlerFunc` that returns responses directly.

## Installation

```bash
go get github.com/sharathcx/fastapify@v0.2.0
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

Handlers use the `fastapify.HandlerFunc` signature (`func(*gin.Context) any`). Return your response directly — Fastapify handles JSON serialization and error formatting automatically.

- Return `fastapify.NewApiResponse(...)` for success responses
- Return `fastapify.NotFound(...)`, `fastapify.BadRequest(...)`, etc. for errors
- Use `fastapify.Req[T](c)` to retrieve the automatically validated and bound request data

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

// UpdateUser - PATCH /users/{id}
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

// DeleteUser - DELETE /users/{id}
func DeleteUser(c *gin.Context) any {
	req := fastapify.Req[UserIdReq](c)

	for i, u := range users {
		if u.ID == req.ID {
			users = append(users[:i], users[i+1:]...)
			return fastapify.NewApiResponse[*User](200, nil, "Deleted")
		}
	}
	return fastapify.NotFound("User not found")
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

	// Setup Swagger UI (JSON at /openapi.json, docs UI at /docs)
	app.SetupSwagger("/openapi.json")

	log.Println("Server running on http://localhost:8080")
	log.Println("Swagger docs available at http://localhost:8080/docs")
	r.Run(":8080")
}
```

## How It Works

### HandlerFunc & Auto-Response

Fastapify uses a custom `HandlerFunc` signature:

```go
type HandlerFunc func(c *gin.Context) any
```

The return value is automatically serialized:
- `*fastapify.ApiError` — formatted as a structured error response with the appropriate status code
- Any other value — serialized as JSON with `200 OK`
- `nil` — no response written (useful if you wrote directly to `c`)

### Auto-Validation Middleware

When you register a route with `.Body(MyStruct{})`, Fastapify automatically injects validation middleware that:

1. Binds URI parameters first and snapshots their values
2. Binds the request body (POST/PUT/PATCH) or query params (GET/DELETE)
3. Restores URI values to prevent body payloads from overriding path parameters
4. Returns a `422` validation error if binding fails
5. Stores the validated struct in the Gin context

Your handler then retrieves the pre-validated data with `fastapify.Req[T](c)` — no manual binding needed.

### Params Binding

For routes with URI parameters, use `.Params()` to declare the params schema separately from the body:

```go
app.GET("/users/{id}", controllers.GetUser).
	Params(controllers.UserIdReq{}).
	Response(controllers.User{})
```

Retrieve params in your handler with `fastapify.Params[T](c)`.

### Route Groups

Group routes under a common prefix:

```go
users := app.Group("/users")

users.GET("/{id}", controllers.GetUser).
	Params(controllers.UserIdReq{}).
	Response(controllers.User{})

users.POST("", controllers.CreateUser).
	Body(controllers.CreateUserReq{}).
	Response(controllers.User{})
```

### Manual Binding

If you prefer to handle binding yourself, you can use `fastapify.Bind()` directly:

```go
func MyHandler(c *gin.Context) any {
	var req MyRequest
	if !fastapify.Bind(c, &req) {
		return nil // error response already sent
	}
	// use req...
	return fastapify.NewApiResponse(200, result, "Success")
}
```

## Error Handling

Fastapify provides convenience constructors for common HTTP errors:

```go
fastapify.NotFound("User not found")
fastapify.BadRequest("Invalid input")
fastapify.Unauthorized("Not authenticated")
fastapify.Forbidden("Access denied")
fastapify.Conflict("Resource already exists")
fastapify.InternalError("Something went wrong")
```

For custom errors:

```go
fastapify.NewApiError(statusCode, "message", fastapify.ErrNotFound, nil)
```

Available error codes: `ErrValidation`, `ErrBadRequest`, `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, `ErrResourceConflict`, `ErrUploadError`, `ErrInternalError`.

### Standardized Response Format

Success responses:

```json
{
  "statusCode": 200,
  "data": { ... },
  "message": "Success",
  "success": true,
  "code": "SUCCESS"
}
```

Error responses:

```json
{
  "success": false,
  "code": "NOT_FOUND",
  "message": "User not found",
  "errors": null
}
```

## Features

- **Custom HandlerFunc:** Return responses directly — Fastapify handles JSON serialization and error formatting.
- **Auto-Validation Middleware:** Requests are automatically validated and bound before your handler runs.
- **URI Parameter Protection:** URI parameters are bound first and protected from being overridden by the request body.
- **Chainable Route Builder:** `app.GET("/path", handler).Body(Req{}).Response(Resp{})` for clean route registration with OpenAPI schema.
- **Params Binding:** Separate URI params from body with `.Params()` and `fastapify.Params[T](c)`.
- **Route Groups:** Group routes under a common prefix with `app.Group("/prefix")`.
- **Auto Swagger Generation:** Inspects your structs and builds an OpenAPI 3.0 spec with Scalar API Reference UI.
- **Flexible Route Syntax:** Supports both Gin-style `:id` and OpenAPI-style `{id}` parameters.
- **Structured Error Handling:** Convenience error constructors and consistent `ApiError` / `ApiResponse` types.
- **Full HTTP Support:** GET, POST, PUT, PATCH, and DELETE methods.
- **Query Parameter Binding:** Use the `form` tag to bind query string parameters.
- **Middleware Support:** Pass per-route middleware: `app.GET("/path", handler, authMiddleware)`.
- **Timeout Middleware:** Built-in `fastapify.TimeoutMiddleware(duration)` for request timeouts.
- **JWT Security Scheme:** OpenAPI spec includes BearerAuth/JWT security scheme by default.
