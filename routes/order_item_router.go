package routes

import (
	"github.com/dastardlyjockey/restaurant-management-backend/controllers"
	"github.com/gin-gonic/gin"
)

func OrderItemRoutes(route *gin.Engine) {
	route.POST("/orderItems", controllers.CreateOrderItem())
	route.GET("/orderItems", controllers.GetOrderItems())
	route.GET("/orderItems/:orderItem_id", controllers.GetOrderItemById())
	route.GET("/orderItems-order/:order_id", controllers.GetOrderItemsByOrder())
	route.PATCH("/orderItems/:orderItem_id", controllers.UpdateOrderItem())
}
