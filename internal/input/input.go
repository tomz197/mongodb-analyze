package input

import (
	"errors"
	"flag"
)

type Flags struct {
	CollectionName string
	DbURI          string
	DbName         string
	PrintJSON      bool
	MaxDepth       *int
}

func ProcessFlags() (*Flags, error) {
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
		CollectionName: *collectionName,
		DbURI:          *dbURI,
		DbName:         *dbName,
		PrintJSON:      *printJSON,
		MaxDepth:       maxDepth,
	}, nil
}
