package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sharathcx/fastapify"
	"github.com/sharathcx/fastapify/examples/user-api/controllers"
)

func main() {
	r := gin.Default()
	app := fastapify.New(r)

	// Register Routes
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
	app.SetupSwagger("/openapi.json")

	log.Println("Server running on http://localhost:8080")
	log.Println("Swagger docs available at http://localhost:8080/docs")
	r.Run(":8080")
}
