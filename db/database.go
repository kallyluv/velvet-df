package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var sess *mongo.Database

// init creates the database connection.
func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017/velvet"))
	if err != nil {
		panic(err)
	}
	sess = client.Database("velvet")
	defer func() {
    	if err = client.Disconnect(ctx); err != nil {
        	panic(err)
    }
}()
}
