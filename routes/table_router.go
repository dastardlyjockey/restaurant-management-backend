package routes

import (
	"github.com/dastardlyjockey/restaurant-management-backend/controllers"
	"github.com/gin-gonic/gin"
)

func TableRoutes(route *gin.Engine) {
	route.POST("/tables", controllers.CreateTable())
	route.GET("/tables", controllers.GetTables())
	route.GET("/tables/:table_id", controllers.GetTableById())
	route.PATCH("/tables/:table_id", controllers.UpdateTable())
}
