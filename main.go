package main

import (
	"fmt"
	"github.com/dastardlyjockey/restaurant-management-backend/database"
	"github.com/dastardlyjockey/restaurant-management-backend/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	defer database.CloseMongoDB(database.Client)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file: ", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	// user routes
	routes.UserRoutes(router)

	// middleware
	router.Use(middleware.Authentication())

	// routes
	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.InvoiceRoutes(router)

	//running server
	fmt.Println("starting server on port: " + port)
	err = router.Run(":" + port)
	if err != nil {
		log.Fatal("server error: ", err)
	}
}
