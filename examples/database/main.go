package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/sharathcx/fastapify"
	"github.com/sharathcx/fastapify/examples/database/controllers"
	"gorm.io/gorm"
)

func main() {
	// 1. Initialize Database
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 2. Pass DB to controllers (Using a global, dependency injection, or struct methods)
	// Here we use a struct to hold our DB reference
	userController := controllers.NewUserController(db)

	// 3. Setup Router
	r := gin.Default()
	app := fastapify.New(r)

	// Register Routes
	fastapify.Get(app, "/users/{id}", userController.GetUser)
	fastapify.Post(app, "/users", userController.CreateUser)
	fastapify.Put(app, "/users/{id}", userController.UpdateUser)
	fastapify.Delete(app, "/users/{id}", userController.DeleteUser)
	fastapify.Get(app, "/users", userController.ListUsers)

	// Setup Swagger UI
	// This will generate docs at /openapi.json and serve the UI at /docs
	app.SetupSwagger("/openapi.json")

	log.Println("Server running on http://localhost:8080")
	log.Println("Swagger docs available at http://localhost:8080/docs")
	r.Run(":8080")
}
