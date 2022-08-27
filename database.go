package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func create_db() (*mongo.Collection, *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	cities := client.Database("Data").Collection("Cities")
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, err := cities.InsertOne(ctx,
			bson.D{
				{"name", 0000},
				{"value", "jgdfsd"},
			},
		)
		log.Print("Successful created cities db")
		log.Print(err)
		return cities, nil
	}
	panic("Trouble while creating database")
}
