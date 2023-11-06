package controllers

import (
	"context"
	"github.com/dastardlyjockey/restaurant-management-backend/database"
	"github.com/dastardlyjockey/restaurant-management-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"time"
)

var tableCollection = database.Collection(database.Client, "table")

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table
		err := c.BindJSON(&table)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse the table"})
			return
		}

		// validate the table
		validateErr := validate.Struct(table)
		if validateErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to validate the table"})
			return
		}

		// populate the table and add it to the database
		table.ID = primitive.NewObjectID()
		table.TableID = table.ID.Hex()
		table.CreatedAt, err = time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
		table.UpdatedAt, err = time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create the time"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, err := tableCollection.InsertOne(ctx, table)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert the table into the database"})
			return
		}

		// success response
		c.JSON(http.StatusCreated, result)
	}
}

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := tableCollection.Find(ctx, bson.D{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get the tables from the database"})
			return
		}

		var allResults []bson.M
		err = cursor.All(ctx, &allResults)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to iterate the tables from the database"})
			return
		}

		c.JSON(http.StatusOK, allResults)
	}
}

func GetTableById() gin.HandlerFunc {
	return func(c *gin.Context) {
		tableId := c.Param("table_id")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var table models.Table

		err := tableCollection.FindOne(ctx, bson.D{{"table_id", tableId}}).Decode(&table)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get the table from the database"})
			return
		}

		c.JSON(http.StatusOK, table)
	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		// send a request
		var table models.Table

		err := c.BindJSON(&table)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse the table"})
			return
		}

		// creating and populate the update object

		var updateObj primitive.D

		if table.TableNumber != nil {
			updateObj = append(updateObj, bson.E{Key: "table_number", Value: table.TableNumber})
		}

		if table.NumberOfGuests != nil {
			updateObj = append(updateObj, bson.E{Key: "number_of_guests", Value: table.NumberOfGuests})
		}

		table.UpdatedAt, err = time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create the time"})
			return
		}

		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: table.UpdatedAt})

		// update the table in the database
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		tableId := c.Param("table_id")
		filter := bson.D{{"table_id", tableId}}

		upsert := true
		updateOptions := options.UpdateOptions{Upsert: &upsert}

		updateResult, err := tableCollection.UpdateOne(ctx, filter, bson.D{{"$set", updateObj}}, &updateOptions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the table in the database"})
			return
		}

		// success response
		c.JSON(http.StatusOK, updateResult)
	}
}
