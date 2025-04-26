package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var uri = os.Getenv("MONGO_URI")

func DBset() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Println("unable to connect to mongodb")
		return nil
	}
	fmt.Println("Mongodb successfully connected")
	return client
}

var Client = DBset()

func SubscriberData(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection = client.Database("Webhook_Delivery").Collection(collectionName)
	return collection
}

func DeliveryData(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection = client.Database("Webhook_Delivery").Collection(collectionName)
	return collection
}
