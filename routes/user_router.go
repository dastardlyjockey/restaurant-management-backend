package routes

import (
	"github.com/dastardlyjockey/restaurant-management-backend/controllers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(route *gin.Engine) {
	route.POST("/users/signup", controllers.Signup)
	route.POST("/users/login", controllers.Login)
	route.GET("/users", controllers.GetUsers())
	route.GET("/users/:user_id", controllers.GetUserById())
}
