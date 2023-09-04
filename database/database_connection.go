package database

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

func DBInstance() *mongo.Client {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading database environment variable")
	}

	db := os.Getenv("DB_URL")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(db))
	if err != nil {
		log.Fatal("error connecting to database")
	}
	//defer func(client *mongo.Client, ctx context.Context) {
	//	err := client.Disconnect(ctx)
	//	if err != nil {
	//		log.Fatal("error disconnecting from database")
	//	}
	//}(client, context.Background())

	fmt.Println("Connected to MongoDB")

	return client
}

var Client = DBInstance()

func Collection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("Restaurant Collection").Collection(collectionName)
	return collection
}

// CloseMongoDB disconnects the MongoDB client when the application shuts down.
func CloseMongoDB(client *mongo.Client) {
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			log.Println("Error disconnecting from database:", err)
		}
		fmt.Println("Disconnected from MongoDB!!")
	}
}
