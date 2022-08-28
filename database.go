package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func create_db() (*mongo.Collection, *mongo.Collection, *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic("Mongo connection failed")
	}
	cities := client.Database("Data").Collection("Cities")
	timers := client.Database("Data").Collection("Timers")
	users := client.Database("Data").Collection("Cities")
	return cities, timers, users
}
