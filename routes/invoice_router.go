package routes

import (
	"github.com/dastardlyjockey/restaurant-management-backend/controllers"
	"github.com/gin-gonic/gin"
)

func InvoiceRoutes(route *gin.Engine) {
	route.POST("/invoices", controllers.CreateInvoice())
	route.GET("/invoices", controllers.GetInvoices())
	route.GET("/invoices/:invoice_id", controllers.GetInvoiceById())
	route.PATCH("/invoices/:invoice_id", controllers.UpdateInvoice())
}
