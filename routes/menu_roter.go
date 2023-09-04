package routes

import (
	"github.com/dastardlyjockey/restaurant-management-backend/controllers"
	"github.com/gin-gonic/gin"
)

func MenuRoutes(route *gin.Engine) {
	route.POST("/menus", controllers.CreateMenu())
	route.GET("/menus", controllers.GetMenus())
	route.GET("/menus/:menu_id", controllers.GetMenuById())
	route.PATCH("/menus/:menu_id", controllers.UpdateMenu())
}
