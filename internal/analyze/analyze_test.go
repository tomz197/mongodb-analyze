package analyze

import (
	"testing"

	"fmt"
	"time"

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
	for _, ts := range stats[name] {
		if ts.GetType() != typ || ts.GetCount() != count {
			return nil, fmt.Errorf("unexpected scalar for %s: type=%s count=%d", name, ts.GetType(), ts.GetCount())
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

func getDocumentStat(stats common.ObjectStats, name string, count int64) (*common.DocumentTypeStat, error) {
	arr, ok := stats[name]
	if !ok {
		return nil, fmt.Errorf("missing stats for %s", name)
	}
	for _, ts := range arr {
		if ds, ok := ts.(*common.DocumentTypeStat); ok {
			if ds.GetCount() != count {
				return nil, fmt.Errorf("unexpected count for %s: %d", name, ds.GetCount())
			}
			return ds, nil
		}
	}
	return nil, fmt.Errorf("no document stat for %s", name)
}

func getArrayStat(stats common.ObjectStats, name string, count int64) (*common.ArrayTypeStat, error) {
	arr, ok := stats[name]
	if !ok {
		return nil, fmt.Errorf("missing stats for %s", name)
	}
	for _, ts := range arr {
		if as, ok := ts.(*common.ArrayTypeStat); ok {
			if as.GetCount() != count {
				return nil, fmt.Errorf("unexpected array count for %s: %d", name, as.GetCount())
			}
			return as, nil
		}
	}
	return nil, fmt.Errorf("no array stat for %s", name)
}

func TestAnalyzeScalarTypes(t *testing.T) {
	dec, err := primitive.ParseDecimal128("123.45")
	if err != nil {
		t.Fatalf("failed to parse decimal128: %v", err)
	}
	doc := bson.D{
		{Key: "str", Value: "Alice"},                                                           // string
		{Key: "dbl", Value: float64(3.14)},                                                     // double
		{Key: "und", Value: primitive.Undefined{}},                                             // undefined
		{Key: "oid", Value: primitive.NewObjectID()},                                           // objectID
		{Key: "bTrue", Value: true},                                                            // bool
		{Key: "bFalse", Value: false},                                                          // bool
		{Key: "date", Value: time.Now()},                                                       // dateTime
		{Key: "null", Value: primitive.Null{}},                                                 // null
		{Key: "regex", Value: primitive.Regex{Pattern: "a", Options: "im"}},                    // regex
		{Key: "dbptr", Value: primitive.DBPointer{DB: "db", Pointer: primitive.NewObjectID()}}, // dbPointer
		{Key: "js", Value: primitive.JavaScript("var x=1;")},                                   // javascript
		{Key: "sym", Value: primitive.Symbol("sym")},                                           // symbol
		{Key: "i32", Value: int32(42)},                                                         // int32
		{Key: "ts", Value: primitive.Timestamp{T: 1, I: 2}},                                    // timestamp
		{Key: "i64", Value: int64(4242)},                                                       // int64
		{Key: "dec128", Value: dec},                                                            // decimal128
		{Key: "minK", Value: primitive.MinKey{}},                                               // minKey
		{Key: "maxK", Value: primitive.MaxKey{}},                                               // maxKey
	}

	elements, root, stats, err := setupTestAnalyzeType(doc)
	if err != nil {
		t.Fatalf("failed to setup test analyze type: %v", err)
	}

	if err := analyze(&root, elements, &stats); err != nil {
		t.Fatalf("analyze returned error: %v", err)
	}

	if _, err = checkTypeStatsAnalyzeType(stats, "str", "string", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "dbl", "double", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "und", "undefined", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "oid", "objectID", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "bTrue", "boolean", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "bFalse", "boolean", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "date", "UTC datetime", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "null", "null", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "regex", "regex", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "dbptr", "dbPointer", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "js", "javascript", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "sym", "symbol", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "i32", "32-bit integer", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "ts", "timestamp", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "i64", "64-bit integer", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "dec128", "128-bit decimal", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "minK", "min key", 1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err = checkTypeStatsAnalyzeType(stats, "maxK", "max key", 1); err != nil {
		t.Fatalf("%v", err)
	}

	if err := checkGlobalStatsAnalyzeType(root, 1, 6, 15, 1); err != nil {
		t.Fatalf("expected global stats: %v", err)
	}
}

func TestLookupBinarySubtypeName_All(t *testing.T) {
	cases := []struct {
		sub byte
		exp string
	}{
		{0x00, "Generic binary subtype"},
		{0x01, "Function"},
		{0x02, "Binary (Old)"},
		{0x03, "UUID (Old)"},
		{0x04, "UUID"},
		{0x05, "MD5"},
		{0x06, "Encrypted BSON value"},
		{0x07, "Compressed BSON column"},
		{0x08, "Sensitive"},
		{0x09, "Vector"},
		{0x80, "User defined"},
	}
	for _, c := range cases {
		if got := lookupBinarySubtypeName(c.sub); got != c.exp {
			t.Fatalf("subtype 0x%02x: got %q want %q", c.sub, got, c.exp)
		}
	}
	if got := lookupBinarySubtypeName(0x99); got != "Unknown" {
		t.Fatalf("expected Unknown for 0x99, got %q", got)
	}
}

func TestElementTypeDisplay_AllBinarySubtypes(t *testing.T) {
	cases := []struct {
		sub byte
		exp string
	}{
		{0x00, "binary - Generic binary subtype"},
		{0x01, "binary - Function"},
		{0x02, "binary - Binary (Old)"},
		{0x03, "binary - UUID (Old)"},
		{0x04, "binary - UUID"},
		{0x05, "binary - MD5"},
		{0x06, "binary - Encrypted BSON value"},
		{0x07, "binary - Compressed BSON column"},
		{0x08, "binary - Sensitive"},
		{0x09, "binary - Vector"},
		{0x80, "binary - User defined"},
	}
	for _, c := range cases {
		doc := bson.D{{Key: "b", Value: primitive.Binary{Subtype: c.sub, Data: []byte{0x01, 0x02}}}}
		rawBytes, err := bson.Marshal(doc)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		raw := bson.Raw(rawBytes)
		elms, err := raw.Elements()
		if err != nil || len(elms) != 1 {
			t.Fatalf("elements error: %v", err)
		}
		if got := elementTypeDisplay(elms[0].Value()); got != c.exp {
			t.Fatalf("subtype 0x%02x display: got %q want %q", c.sub, got, c.exp)
		}
	}
}

func TestElementTypeDisplay_BinaryAndScalar(t *testing.T) {
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

func TestAnalyze_NestedDocumentsAndArrays(t *testing.T) {
	doc := bson.D{
		{Key: "profile", Value: bson.D{
			{Key: "name", Value: "Alice"},
			{Key: "age", Value: int32(30)},
			{Key: "meta", Value: bson.D{
				{Key: "active", Value: true},
				{Key: "score", Value: float64(9.5)},
			}},
		}},
		{Key: "tags", Value: bson.A{"go", 1, true}},
		{Key: "nums", Value: bson.A{int32(1), int64(2), float64(3)}},
		{Key: "docs", Value: bson.A{
			bson.D{{Key: "k", Value: "v1"}},
			bson.D{{Key: "k", Value: "v2"}},
		}},
	}

	elements, root, stats, err := setupTestAnalyzeType(doc)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	if err := analyze(&root, elements, &stats); err != nil {
		t.Fatalf("analyze error: %v", err)
	}

	if root.Depth < 2 {
		t.Fatalf("expected depth >= 2, got %d", root.Depth)
	}

	prof, err := getDocumentStat(stats, "profile", 1)
	if err != nil {
		t.Fatalf("profile stat: %v", err)
	}
	if _, err := checkTypeStatsAnalyzeType(prof.Props, "name", "string", 1); err != nil {
		t.Fatalf("profile.name: %v", err)
	}
	if _, err := checkTypeStatsAnalyzeType(prof.Props, "age", "32-bit integer", 1); err != nil {
		t.Fatalf("profile.age: %v", err)
	}
	meta, err := getDocumentStat(prof.Props, "meta", 1)
	if err != nil {
		t.Fatalf("profile.meta: %v", err)
	}
	if _, err := checkTypeStatsAnalyzeType(meta.Props, "active", "boolean", 1); err != nil {
		t.Fatalf("meta.active: %v", err)
	}
	if _, err := checkTypeStatsAnalyzeType(meta.Props, "score", "double", 1); err != nil {
		t.Fatalf("meta.score: %v", err)
	}

	tags, err := getArrayStat(stats, "tags", 1)
	if err != nil {
		t.Fatalf("tags stat: %v", err)
	}
	if tags.Items["string"] != 1 || tags.Items["32-bit integer"] != 1 || tags.Items["boolean"] != 1 {
		t.Fatalf("unexpected tags item counts: %+v", tags.Items)
	}

	nums, err := getArrayStat(stats, "nums", 1)
	if err != nil {
		t.Fatalf("nums stat: %v", err)
	}
	if nums.Items["32-bit integer"] != 1 || nums.Items["64-bit integer"] != 1 || nums.Items["double"] != 1 {
		t.Fatalf("unexpected nums item counts: %+v", nums.Items)
	}

	docs, err := getArrayStat(stats, "docs", 1)
	if err != nil {
		t.Fatalf("docs stat: %v", err)
	}
	if docs.ItemProps == nil {
		t.Fatalf("docs.ItemProps nil")
	}
	if _, err := checkTypeStatsAnalyzeType(*docs.ItemProps, "k", "string", 2); err != nil {
		t.Fatalf("docs.k: %v", err)
	}
}
