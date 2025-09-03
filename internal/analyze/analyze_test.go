package analyze

import (
	"testing"

	"fmt"

	"github.com/tomz197/mongodb-analyze/internal/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupTestAnalyzeType(doc bson.D) ([]bson.RawElement, common.RootObject, common.ObjectStats, error) {
	root := common.RootObject{Depth: 0, MaxDepth: nil, MaxNameLen: 0, MaxTypeLen: 0, TotalObjects: 1, Stats: common.ObjectStats{}}
	stats := common.ObjectStats{}

	rawBytes, err := bson.Marshal(doc)
	if err != nil {
		return nil, root, stats, fmt.Errorf("failed to marshal doc: %v", err)
	}

	raw := bson.Raw(rawBytes)
	elements, err := raw.Elements()
	if err != nil {
		return nil, root, stats, fmt.Errorf("failed to get elements: %v", err)
	}

	return elements, root, stats, nil
}

func checkTypeStatsAnalyzeType(stats common.ObjectStats, name string, typ string, count int64) (common.TypeStat, error) {
	if _, ok := stats[name]; !ok {
		return nil, fmt.Errorf("expected stats for '%s'", name)
	}
	var found common.TypeStat = nil
	for _, ts := range stats["name"] {
		if ts.GetType() != typ || ts.GetCount() != count {
			return nil, fmt.Errorf("unexpected scalar for name: type=%s count=%d", ts.GetType(), ts.GetCount())
		}
		found = ts
	}
	if found == nil {
		return nil, fmt.Errorf("expected scalar %s type for '%s'", typ, name)
	}
	return found, nil
}

func checkGlobalStatsAnalyzeType(root common.RootObject, depth int, nameLen int, typeLen int, totalObjects int64) error {
	if root.Depth != depth {
		return fmt.Errorf("expected Depth %d, got %d", depth, root.Depth)
	}
	if root.MaxNameLen < nameLen {
		return fmt.Errorf("expected MaxNameLen >= %d, got %d", nameLen, root.MaxNameLen)
	}
	if root.MaxTypeLen < typeLen {
		return fmt.Errorf("expected MaxTypeLen >= %d, got %d", typeLen, root.MaxTypeLen)
	}
	if root.TotalObjects != totalObjects {
		return fmt.Errorf("expected TotalObjects %d, got %d", totalObjects, root.TotalObjects)
	}
	return nil
}

func TestAnalyzeScalarTypes(t *testing.T) {
	doc := bson.D{
		{Key: "name", Value: "Alice"}, // string
	}

	elements, root, stats, err := setupTestAnalyzeType(doc)
	if err != nil {
		t.Fatalf("failed to setup test analyze type: %v", err)
	}

	if err := analyze(&root, elements, &stats); err != nil {
		t.Fatalf("analyze returned error: %v", err)
	}

	_, err = checkTypeStatsAnalyzeType(stats, "name", "string", 1)
	if err != nil {
		t.Fatalf("expected stats for 'name'")
	}

	if err := checkGlobalStatsAnalyzeType(root, 1, len("name"), len("string"), 1); err != nil {
		t.Fatalf("expected global stats: %v", err)
	}
}

func TestElementTypeDisplay_BinaryAndScalar(t *testing.T) {
	// Build a small document to extract RawValue
	doc := bson.D{
		{Key: "x", Value: primitive.Binary{Subtype: 0x04, Data: []byte{0x01}}},
		{Key: "y", Value: "str"},
	}
	rawBytes, err := bson.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	raw := bson.Raw(rawBytes)
	elms, err := raw.Elements()
	if err != nil || len(elms) < 2 {
		t.Fatalf("failed to get elements: %v", err)
	}
	if got := elementTypeDisplay(elms[0].Value()); got != "binary - UUID" {
		t.Fatalf("unexpected elementTypeDisplay for binary: %s", got)
	}
	if got := elementTypeDisplay(elms[1].Value()); got != "string" {
		t.Fatalf("unexpected elementTypeDisplay for string: %s", got)
	}
}

func TestLookupBinarySubtypeName(t *testing.T) {
	if got := lookupBinarySubtypeName(0x04); got != "UUID" {
		t.Fatalf("expected UUID, got %s", got)
	}
	if got := lookupBinarySubtypeName(0x99); got != "Unknown" {
		t.Fatalf("expected Unknown, got %s", got)
	}
}

func TestDescendIntoDocument_NonEmbedded(t *testing.T) {
	// Create a scalar value and ensure descendIntoDocument ignores it without error
	doc := bson.D{{Key: "x", Value: "str"}}
	rawBytes, err := bson.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	raw := bson.Raw(rawBytes)
	elms, err := raw.Elements()
	if err != nil || len(elms) == 0 {
		t.Fatalf("failed to get elements: %v", err)
	}
	root := &common.RootObject{}
	props, derr := descendIntoDocument(root, elms[0].Value(), nil)
	if derr != nil {
		t.Fatalf("unexpected error: %v", derr)
	}
	if props != nil {
		t.Fatalf("expected nil props for non-embedded value")
	}
}
