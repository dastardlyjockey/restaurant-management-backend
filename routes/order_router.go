package routes

import (
	"github.com/dastardlyjockey/restaurant-management-backend/controllers"
	"github.com/gin-gonic/gin"
)

func OrderRoutes(route *gin.Engine) {
	route.POST("/orders", controllers.CreateOrder())
	route.GET("/orders", controllers.GetOrders())
	route.GET("/orders/:order_id", controllers.GetOrderById())
	route.PATCH("/orders/:order_id", controllers.UpdateOrder())
}
