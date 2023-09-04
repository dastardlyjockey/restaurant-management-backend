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

var menuCollection = database.Collection(database.Client, "menu")

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}

func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the food request from the body
		var menu models.Menu

		err := c.BindJSON(&menu)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu JSON"})
			return
		}

		// validate the food structure
		err = validate.Struct(menu)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "the validation of the menu structure failed"})
			return
		}

		// insert into the database
		menu.CreatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.UpdatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			log.Println("Failed in creating food timestamp")
		}
		menu.ID = primitive.NewObjectID()
		menu.MenuID = menu.ID.Hex()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		result, err := foodCollection.InsertOne(ctx, menu)
		if err != nil {
			msg := fmt.Sprintf("Menu was not created in the database")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		// response for successful food item insertion
		c.JSON(http.StatusOK, result)
	}
}

func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := menuCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing the menus"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var allMenus []bson.M
		if err = result.All(ctx, &allMenus); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allMenus)
	}
}

func GetMenuById() gin.HandlerFunc {
	return func(c *gin.Context) {
		menuId := c.Param("menu_id")

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var menu models.Menu

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching the menu in the database"})
			return
		}

		c.JSON(http.StatusOK, menu)
	}
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get menu from the request body
		var menu models.Menu

		err := c.BindJSON(menu)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while binding the menu JSON from the request body"})
			return
		}

		// get the menuID from the url parameter
		menuId := c.Param("menu_id")
		filter := bson.M{"menu_id": menuId}

		// create the update obj to be used in the menu collection
		var updateObj primitive.D

		//run check to see if the details are inputted correctly
		if menu.StartDate != nil && menu.EndDate != nil {
			if !inTimeSpan(*menu.StartDate, *menu.EndDate, time.Now()) {
				msg := fmt.Sprintf("Kindly retype the time")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}

			updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.StartDate})
			updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.EndDate})

			if menu.Name != "" {
				updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
			}

			if menu.Category != "" {
				updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
			}

			menu.UpdatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing the updated time"})
				return
			}

			updateObj = append(updateObj, bson.E{Key: "updated_at", Value: menu.UpdatedAt})

			//update the menu collection
			upsert := true
			opt := options.UpdateOptions{Upsert: &upsert}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			result, err := menuCollection.UpdateOne(ctx, filter, bson.D{{"$set", updateObj}}, &opt)
			if err != nil {
				msg := fmt.Sprintf("menu failed to update")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}

			// a successful response
			c.JSON(http.StatusOK, result)
		}
	}
}
