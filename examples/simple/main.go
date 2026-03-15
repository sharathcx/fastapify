package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sharathcx/fastapify"
	"github.com/sharathcx/fastapify/examples/simple/controllers"
)

func main() {
	r := gin.Default()
	app := fastapify.New(r)

	// Register Routes
	fastapify.Get(app, "/users/{id}", controllers.GetUser)
	fastapify.Post(app, "/users", controllers.CreateUser)
	fastapify.Put(app, "/users/{id}", controllers.UpdateUser)
	fastapify.Delete(app, "/users/{id}", controllers.DeleteUser)

	// Setup Swagger UI
	// This will generate docs at /openapi.json and serve the UI at /docs
	app.SetupSwagger("/openapi.json")

	log.Println("Server running on http://localhost:8080")
	log.Println("Swagger docs available at http://localhost:8080/docs")
	r.Run(":8080")
}
