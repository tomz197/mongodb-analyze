package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetCollection(ctx context.Context, uri, dbName, collectionName string) (*mongo.Client, *mongo.Collection, error) {
	client, err := connectToMongoDB(ctx, uri)
	if err != nil {
		return nil, nil, err
	}

	// Get collection
	collection := client.Database(dbName).Collection(collectionName)

	return client, collection, nil
}

func connectToMongoDB(ctx context.Context, dbURI string) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}
