package controllers

import (
	"context"
	"fmt"
	"github.com/dastardlyjockey/restaurant-management-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"strconv"
	"time"
)

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, _ = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		ProjectStage := bson.D{{"$project", bson.D{
			{"_id", 0},
			{"total_count", 1},
			{"user_items", bson.D{
				{"$slice", []interface{}{"$data", startIndex, recordPerPage}},
			}},
		}}}

		// create a database aggregation
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{matchStage, ProjectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error occurred while aggregating the database: %v", err.Error())})
			return
		}

		var allUsers []bson.M
		err = cursor.All(ctx, &allUsers)
		if err != nil {
			log.Fatal(err)
		}

		//response
		c.JSON(http.StatusOK, allUsers)

	}
}

func GetUserById() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		err := UserCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve the user"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
