package routes

import (
	"github.com/dastardlyjockey/restaurant-management-backend/controllers"
	"github.com/gin-gonic/gin"
)

func FoodRoutes(route *gin.Engine) {
	route.POST("/foods", controllers.CreateFood())
	route.GET("/foods", controllers.GetFoods())
	route.GET("/foods/:food_id", controllers.GetFoodById())
	route.PATCH("/foods/:food_id", controllers.UpdateFood())
}
