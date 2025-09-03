package main

import (
	"context"
	"fmt"
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

	file := os.Stdout
	if flags.OutputFile != nil {
		file, err = os.Create(*flags.OutputFile)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer func() {
			fmt.Println("Result saved to", *flags.OutputFile)
		}()
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("failed to close output file: %v", err)
			}
		}()
	}

	root, err := analyze.All(analyze.AllOptions{
		Context:    ctx,
		Collection: collection,
		MaxDepth:   flags.MaxDepth,
	})
	if err != nil {
		log.Fatalf("failed to analyze collection: %v", err)
	}

	if flags.PrintJSON {
		output.PrintJSON(root, file)
	} else {
		output.PrintTable(root, file)
	}

}
