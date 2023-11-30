package controllers

import (
	"context"
	"github.com/dastardlyjockey/restaurant-management-backend/database"
	"github.com/dastardlyjockey/restaurant-management-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

var orderItemsCollection = database.Collection(database.Client, "orderItems")

type orderItemPack struct {
	TableID    *string
	OrderItems []models.OrderItem
}

func ItemsByOrder(id string) (orderItems []primitive.M, err error) {
	matchStage := bson.D{{"$match", bson.D{{"order_id", id}}}}
	lookupStage := bson.D{{"$lookup", bson.D{{"from", "order"}, {"localField", "order_id"}, {"foreignField", "order_id"}, {"as", "order"}}}}
	unwindStage := bson.D{{"$unwind", bson.D{{"path", "$order"}, {"preserveNullAndEmptyArrays", true}}}}
	lookupFoodStage := bson.D{{"$lookup", bson.D{{"from", "food"}, {"localField", "food_id"}, {"foreignField", "food_id"}, {"as", "food"}}}}
	unwindFoodStage := bson.D{{"$unwind", bson.D{{"path", "$food"}, {"preserveNullAndEmptyArrays", true}}}}
	lookupTableStage := bson.D{{"$lookup", bson.D{{"from", "table"}, {"localField", "order.table_id"}, {"foreignField", "table_id"}, {"as", "table"}}}}
	unwindTableStage := bson.D{{"$unwind", bson.D{{"path", "$table"}, {"preserveNullAndEmptyArrays", true}}}}

	projectStage := bson.D{{"$project", bson.D{
		{"_id", 0},
		{"amount", "$food.price"},
		{"total_count", 1},
		{"food_name", "$food.name"},
		{"food_image", "$food.food_image"},
		{"table_number", "$table.table_number"},
		{"table_id", "$table.table_id"},
		{"order_id", "$order.order_id"},
		{"price", "$food.price"},
		{"quantity", 1},
	}}}

	groupStage := bson.D{{"$group", bson.D{
		{"_id", bson.D{{"order_id", "$order_id"}, {"table_id", "$table_id"}, {"table_number", "$table_number"}}},
		{"payment_due", bson.D{{"$sum", "$amount"}}},
		{"total_count", bson.D{{"$sum", 1}}},
		{"order_items", bson.D{{"$push", "$order_items"}}},
	}}}

	secondProjectStage := bson.D{{"$project", bson.D{
		{"_id", 0},
		{"payment_due", 1},
		{"total_count", 1},
		{"table_number", "$_id.table_number"},
		{"order_items", 1},
	}}}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	cursor, err := orderItemsCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookupStage,
		unwindStage,
		lookupFoodStage,
		unwindFoodStage,
		lookupTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		secondProjectStage,
	})
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &orderItems)
	if err != nil {
		return nil, err
	}

	return orderItems, nil
}

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the request from the client
		var orderItemPack orderItemPack
		var order models.Order
		err := c.BindJSON(&orderItemPack)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to parse the order items"})
			return
		}

		// add the order item to the database
		orderItemToBeInserted := []interface{}{}
		order.TableID = orderItemPack.TableID
		orderID := OrderItemOrderCreator(order)

		for _, orderItem := range orderItemPack.OrderItems {
			orderItem.OrderID = orderID

			validateErr := validate.Struct(orderItem)
			if validateErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid order item"})
				return
			}

			orderItem.ID = primitive.NewObjectID()
			orderItem.OrderItemID = orderItem.ID.Hex()
			num := toFixed(*orderItem.UnitPrice, 2)
			orderItem.UnitPrice = &num
			orderItem.CreatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.CreatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create time"})
				return
			}

			orderItemToBeInserted = append(orderItemToBeInserted, orderItem)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, err := orderItemsCollection.InsertMany(ctx, orderItemToBeInserted)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "failed to insert into database"})
			return
		}

		//response
		c.JSON(http.StatusOK, result)

	}
}

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetOrderItemById() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
