package analyze

import (
	"context"
	"fmt"

	"github.com/tomz197/mongodb-analyze/internal/common"
	"github.com/tomz197/mongodb-analyze/internal/output"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AllOptions struct {
	// The context to use for the operation
	Context context.Context
	// The collection to query
	Collection *mongo.Collection
	// The maximum depth to analyze
	MaxDepth *int
}

func All(options AllOptions) (*common.RootObject, error) {
	estimatedCount, err := options.Collection.EstimatedDocumentCount(options.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get estimated document count: %v", err)
	}

	// Find all documents
	cursor, err := options.Collection.Find(options.Context, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(options.Context); err != nil {
			fmt.Println("failed to close cursor:", err)
		}
	}()

	root := &common.RootObject{
		Depth:        0,
		MaxDepth:     options.MaxDepth,
		MaxNameLen:   20,
		MaxTypeLen:   20,
		TotalObjects: 0,
		Stats:        common.ObjectStats{},
	}

	printProgress, finished := output.GetPrintProgress(estimatedCount)

	fmt.Println("\nEstimated object count:", estimatedCount)

	for cursor.Next(options.Context) {
		document := cursor.Current
		root.TotalObjects++
		printProgress(root.TotalObjects)

		elements, err := document.Elements()
		if err != nil {
			return root, fmt.Errorf("failed to get values from document: %v", err)
		}

		err = analyze(root, elements, &root.Stats)
		if err != nil {
			return root, fmt.Errorf("failed to analyze document: %v", err)
		}
	}
	finished(root.TotalObjects)

	if cursor.Err() != nil {
		return root, fmt.Errorf("cursor error: %v", cursor.Err())
	}

	return root, nil
}
