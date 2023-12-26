package controllers

import (
	"context"
	"github.com/dastardlyjockey/restaurant-management-backend/database"
	"github.com/dastardlyjockey/restaurant-management-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := orderItemsCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while listing the order items"})
			return
		}

		var allOrder []bson.D
		err = cursor.All(ctx, &allOrder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, allOrder)
	}
}

func GetOrderItemById() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderItemId := c.Param("orderItem_id")

		var orderItem models.OrderItem

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := orderItemsCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, orderItem)
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderId := c.Param("order_id")

		allItemsByOrder, err := ItemsByOrder(orderId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get items by order"})
			return
		}

		c.JSON(http.StatusOK, allItemsByOrder)
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var orderItem models.OrderItem
		orderItemId := c.Param("orderItem_id")
		filter := bson.M{"order_item_id": orderItemId}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var updateObj primitive.D

		if orderItem.UnitPrice != nil {
			updateObj = append(updateObj, bson.E{"unit_price", *orderItem.UnitPrice})
		}

		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{"quantity", *orderItem.Quantity})
		}

		if orderItem.FoodID != nil {
			updateObj = append(updateObj, bson.E{"food_id", *orderItem.FoodID})
		}

		orderItem.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", orderItem.UpdatedAt})

		//update the order item in the database
		upsert := true
		opt := options.UpdateOptions{Upsert: &upsert}

		result, err := orderItemsCollection.UpdateOne(ctx, filter, bson.D{{"$set", updateObj}}, &opt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the order item"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
