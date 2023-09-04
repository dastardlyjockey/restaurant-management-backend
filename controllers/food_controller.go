package controllers

import (
	"context"
	"fmt"
	"github.com/dastardlyjockey/restaurant-management-backend/database"
	"github.com/dastardlyjockey/restaurant-management-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

var foodCollection = database.Collection(database.Client, "food")

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the food request from the body
		var food models.Food
		var menu models.Menu

		err := c.BindJSON(&food)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food JSON"})
			return
		}

		// validate the food structure
		err = validate.Struct(food)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "the validation of the food structure failed"})
			return
		}

		// insert into the database
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		err = menuCollection.FindOne(ctx, bson.M{"menu_id": food.MenuID}).Decode(&menu)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "menu was not found"})
			return
		}

		food.CreatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.UpdatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			log.Println("Failed in creating food timestamp")
		}
		food.ID = primitive.NewObjectID()
		food.FoodID = food.ID.Hex()

		num := toFixed(*food.Price, 2)
		food.Price = &num

		result, err := foodCollection.InsertOne(ctx, food)
		if err != nil {
			msg := fmt.Sprintf("Food item was not created in the database")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		// response for successful food item insertion
		c.JSON(http.StatusOK, result)

	}
}

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("Page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}},
			{"total_count", bson.D{{"$sum", 1}}},
			{"data", bson.D{{"$push", "$$ROOT"}}},
		}}}
		projectStage := bson.D{{"$project", bson.D{
			{"_id", 0},
			{"total_count", 1},
			{"food_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
		}}}

		// get data from the database
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed in retrieving the aggregated data"})
			return
		}

		var allFood []bson.M
		err = result.All(ctx, &allFood)
		if err != nil {
			log.Fatal(err)
			return
		}

		// response
		c.JSON(http.StatusOK, allFood[0])
	}
}

func GetFoodById() gin.HandlerFunc {
	return func(c *gin.Context) {
		foodId := c.Param("food_id")

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var food models.Food

		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching the food in the database"})
			return
		}

		c.JSON(http.StatusOK, food)
	}
}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get request from the body
		var food models.Food

		err := c.BindJSON(&food)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while binding the food JSON from the request body"})
			return
		}

		// retrieve the food id you want to update
		foodID := c.Param("food_id")
		filter := bson.M{"food_id": foodID}

		// create the update Obj
		var updateObj primitive.D

		// verify the data from the food and store it in the update obj

		if *food.Name != "" {
			updateObj = append(updateObj, bson.E{Key: "name", Value: *food.Name})
		}

		if food.Price != nil {
			num := toFixed(*food.Price, 2)
			food.Price = &num
			updateObj = append(updateObj, bson.E{Key: "price", Value: *food.Price})
		}

		if food.FoodImage != nil {
			updateObj = append(updateObj, bson.E{Key: "food_image", Value: *food.FoodImage})
		}

		//check whether the food belongs to a menu
		var menu models.Menu

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		err = menuCollection.FindOne(ctx, bson.M{"menu_id": food.MenuID}).Decode(&menu)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "menu was not found"})
			return
		}

		// update the time
		food.UpdatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			log.Println("Failed in updating food timestamp")
		}

		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.UpdatedAt})

		// update the food id in the database
		upsert := true
		opt := options.UpdateOptions{Upsert: &upsert}

		updateResult, err := foodCollection.UpdateOne(ctx, filter, bson.D{{"$set", updateObj}}, &opt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed updating food"})
			return
		}

		// response
		c.JSON(http.StatusOK, updateResult)
	}
}
