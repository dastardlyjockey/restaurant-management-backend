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

type InvoiceViewFormat struct {
	InvoiceID      string
	PaymentMethod  string
	OrderID        string
	PaymentStatus  *string
	PaymentDue     interface{}
	TableNumber    interface{}
	PaymentDueDate time.Time
	OrderDetails   interface{}
}

var invoiceCollection = database.Collection(database.Client, "invoice")

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get request from the body
		var invoice models.Invoice

		err := c.BindJSON(&invoice)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice JSON"})
			return
		}

		// validate the invoice

		err = validate.Struct(invoice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invoice validation failed"})
			return
		}

		// check if the order ID exists
		var order models.Order

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = orderCollection.FindOne(ctx, bson.M{"order_id": invoice.OrderID}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "order is not found"})
			return
		}

		// add the invoice to the database
		invoice.ID = primitive.NewObjectID()
		invoice.InvoiceID = invoice.ID.Hex()

		invoice.CreatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.UpdatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			log.Println("Error creating timestamp: ", err)
			return
		}

		result, err := invoiceCollection.InsertOne(ctx, &invoice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
			return
		}

		// response
		c.JSON(http.StatusOK, result)
	}
}

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		cursor, err := invoiceCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			msg := fmt.Sprintf("Error getting the invoices, error:%s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		var results []bson.M
		err = cursor.All(context.TODO(), &results)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store the invoice results"})
			return
		}

		c.JSON(http.StatusOK, results)
	}
}

func GetInvoiceById() gin.HandlerFunc {
	return func(c *gin.Context) {
		//Get the invoice id
		invoiceId := c.Param("invoice_id")

		//search for the invoice in the database
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var invoice models.Invoice
		err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "The invoice is not in the database"})
			return
		}

		// integrate the invoice into my created format
		var invoiceView InvoiceViewFormat

		//orderItems, err := ItemsByOrder(invoice.OrderID)
		//if err != nil {
		//	log.Println("error retrieving the Items by order", err)
		//	return
		//}

		invoiceView.OrderID = invoice.OrderID
		invoiceView.InvoiceID = invoice.InvoiceID
		invoiceView.PaymentDueDate = invoice.PaymentDueDate
		invoiceView.PaymentMethod = "Null"
		if invoice.PaymentMethod != nil {
			invoiceView.PaymentMethod = *invoice.PaymentMethod
		}
		invoiceView.PaymentStatus = *&invoice.PaymentStatus
		//invoiceView.PaymentDue = orderItems[0]{"payment_due"}
		//invoiceView.OrderDetails = orderItems[0]{"order_details"}
		//invoiceView.TableNumber = orderItems[0]{"table_number"}

		// response
		c.JSON(http.StatusOK, invoiceView)
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the invoice request from the body
		var invoice models.Invoice

		err := c.BindJSON(&invoice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse the update invoice"})
			return
		}

		//validate the invoice
		err = validate.Struct(invoice)
		if err != nil {
			c.JSON(http.StatusNotAcceptable, gin.H{"error": "Validation failed"})
			return
		}

		// creating the update object and retrieving the invoice id
		invoiceId := c.Param("invoice_id")
		filter := bson.M{"invoice_id": invoiceId}

		var updateObj primitive.D

		// add the update to the database
		if invoice.PaymentMethod != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_method", Value: invoice.PaymentMethod})
		}

		if invoice.PaymentStatus != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_status", Value: invoice.PaymentStatus})
		}

		status := "PENDING"
		if invoice.PaymentStatus == nil {
			invoice.PaymentStatus = &status
			updateObj = append(updateObj, bson.E{Key: "payment_status", Value: invoice.PaymentStatus})
		}

		upsert := true
		obj := options.UpdateOptions{Upsert: &upsert}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, err := invoiceCollection.UpdateOne(ctx, filter, bson.D{{"$set", updateObj}}, &obj)
		if err != nil {
			msg := fmt.Sprintf("Error updating the invoice database: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		//response
		c.JSON(http.StatusOK, result)
	}
}
