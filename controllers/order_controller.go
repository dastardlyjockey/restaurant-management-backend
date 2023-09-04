package controllers

import (
	"context"
	"fmt"
	"github.com/dastardlyjockey/restaurant-management-backend/database"
	"github.com/dastardlyjockey/restaurant-management-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"time"
)

var orderCollection = database.Collection(database.Client, "orders")

func OrderItemOrderCreator(order models.Order) string {

}

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		//Get the request from the body
		var order models.Order

		err := c.BindJSON(&order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to bind the order JSON"})
			return
		}

		// validate the order

	}
}

func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := orderCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve the orders from the database"})
			return
		}

		var allResult []bson.M
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = result.All(ctx, &allResult)
		if err != nil {
			log.Fatal("Failed to iterate the orders")
			return
		}

		c.JSON(http.StatusOK, allResult)
	}
}

func GetOrderById() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderId := c.Param("order_id")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var order models.Order
		err := orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
		if err != nil {
			msg := fmt.Sprintf("Error retrieving the order: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
