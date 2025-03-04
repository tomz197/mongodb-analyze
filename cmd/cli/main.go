package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tomz197/mongodb-analyze/internal/analyze"
	"github.com/tomz197/mongodb-analyze/internal/database"
	"github.com/tomz197/mongodb-analyze/internal/output"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Flags struct {
	collectionName string
	dbURI          string
	dbName         string
	printJSON      bool
	maxDepth       *int
}

func main() {
	flags, err := handleFlags()
	if err != nil {
		log.Fatalf("Failed to handle flags: %v", err)
		os.Exit(1)
	}

	// Set up MongoDB client
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := database.ConnectToMongoDB(ctx, flags.dbURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
		return
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	fmt.Println("Analyzing collection:", flags.collectionName)

	// Get collection
	collection := client.Database(flags.dbName).Collection(flags.collectionName)
	estimatedCount, err := collection.EstimatedDocumentCount(ctx)

	// Find all documents
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to query collection: %v", err)
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error: %v", err)
		return
	}

	root, err := analyzeTypes(ctx, flags, cursor, estimatedCount)
	if err != nil {
		log.Fatalf("Failed to analyze records: %v", err)
		return
	}

	if flags.printJSON {
		output.PrintJSON(root)
	} else {
		output.PrintTable(root)
	}

}

func handleFlags() (*Flags, error) {
	// Define command line flags
	collectionName := flag.String("collection", "", "Name of the MongoDB collection (required)")
	dbURI := flag.String("uri", "mongodb://localhost:27017", "MongoDB connection URI")
	dbName := flag.String("db", "test", "MongoDB database name")
	printJSON := flag.Bool("json", false, "Print output stats in JSON")
	maxDepth := flag.Int("depth", 0, "Maximum depth to analyze (0 for all)")

	// Parse command line flags
	flag.Parse()

	// Check if collection name was provided
	if *collectionName == "" {
		flag.Usage()
		return nil, errors.New("missing collection name")
	}

	if *maxDepth == 0 {
		maxDepth = nil
	}

	return &Flags{
		collectionName: *collectionName,
		dbURI:          *dbURI,
		dbName:         *dbName,
		printJSON:      *printJSON,
		maxDepth:       maxDepth,
	}, nil
}

func analyzeTypes(ctx context.Context, flags *Flags, cursor *mongo.Cursor, estimatedCount int64) (root *analyze.RootObject, err error) {
	root = &analyze.RootObject{
		Depth:        0,
		MaxDepth:     flags.maxDepth,
		NameLens:     []int{},
		TotalObjects: 0,
		Stats:        analyze.ObjectStats{},
	}

	percOfDocs := estimatedCount / 100

	fmt.Println("\nEstimated object count:", estimatedCount)

	for cursor.Next(ctx) {
		var document bson.Raw = cursor.Current
		root.TotalObjects++
		if root.TotalObjects%percOfDocs == 0 {
			bar := ""
			for range root.TotalObjects / (percOfDocs * 5) {
				bar += "="
			}
			if len(bar) < 20 {
				bar += ">"
			}
			fmt.Printf("\r Progress [%-20s] %d%% (%d/%d)", bar, root.TotalObjects/percOfDocs, root.TotalObjects, estimatedCount)
		}

		elements, err := document.Elements()
		if err != nil {
			return root, fmt.Errorf("Failed to get values from document: %v", err)
		}

		err = analyze.AnalyzeBSON(root, elements, &root.Stats)
		if err != nil {
			return root, fmt.Errorf("Failed to analyze document: %v", err)
		}
	}
	fmt.Println()

	if cursor.Err() != nil {
		return root, fmt.Errorf("Cursor error: %v", cursor.Err())
	}

	return root, nil
}
