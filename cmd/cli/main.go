package main

import (
	"context"
	"log"
	"os"

	"github.com/tomz197/mongodb-analyze/internal/analyze"
	"github.com/tomz197/mongodb-analyze/internal/database"
	"github.com/tomz197/mongodb-analyze/internal/input"
	"github.com/tomz197/mongodb-analyze/internal/output"
)

func main() {
	flags, err := input.ProcessFlags()
	if err != nil {
		log.Fatalf("Failed to handle flags: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, collection, err := database.GetCollection(ctx, flags.DbURI, flags.DbName, flags.CollectionName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
		return
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	root, err := analyze.All(analyze.AllOptions{
		Context:    ctx,
		Collection: collection,
		MaxDepth:   flags.MaxDepth,
	})

	if flags.PrintJSON {
		output.PrintJSON(root)
	} else {
		output.PrintTable(root)
	}

}
