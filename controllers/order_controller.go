package controllers

import (
	"context"
	"fmt"
	"github.com/dastardlyjockey/restaurant-management-backend/database"
	"github.com/dastardlyjockey/restaurant-management-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

var orderCollection = database.Collection(database.Client, "orders")

func OrderItemOrderCreator(order models.Order) string {
	order.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.OrderID = order.ID.Hex()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	orderCollection.InsertOne(ctx, order)
	return order.OrderID
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
		err = validate.Struct(order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate order JSON"})
			return
		}

		//check if the table id is valid
		var table models.Table
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = tableCollection.FindOne(ctx, bson.M{"table_id": order.TableID}).Decode(&table)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The table Id is not available"})
			return
		}

		// input the data to the database
		order.CreatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.UpdatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			msg := fmt.Sprintf("Failed to create timestamp, error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		order.ID = primitive.NewObjectID()
		order.OrderID = order.ID.Hex()

		result, err := orderCollection.InsertOne(ctx, order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert the order into the database"})
			return
		}

		// message for a successful request
		c.JSON(http.StatusCreated, result)
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
		// get request from the body
		var order models.Order

		err := c.BindJSON(&order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while binding the order JSON from the request body"})
			return
		}

		// retrieve the order id you want to update
		orderID := c.Param("order_id")
		filter := bson.M{"order_id": orderID}

		// create the update Obj
		var updateObj primitive.D

		// verify the data from the order and store it in the update obj

		//check whether the order belongs to a table
		var table models.Table

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		err = tableCollection.FindOne(ctx, bson.M{"table_id": order.TableID}).Decode(&table)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Table was not found"})
			return
		}

		// update the time
		order.UpdatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			log.Println("Failed in updating timestamp")
		}

		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: order.UpdatedAt})

		// update the order in the database
		upsert := true
		opt := options.UpdateOptions{Upsert: &upsert}

		updateResult, err := orderCollection.UpdateOne(ctx, filter, bson.D{{"$set", updateObj}}, &opt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed updating order"})
			return
		}

		// response
		c.JSON(http.StatusOK, updateResult)
	}
}
