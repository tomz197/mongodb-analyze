package analyze

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"
)

// Custom errors
var (
	ErrInvalidDocument = errors.New("invalid document")
)

// ObjectStats maps field names to their type statistics
type ObjectStats map[string][]TypeStats

// RootObject is has more statistics/flags
type RootObject struct {
	CurrDepth    int   // NOTE: current depth of the document
	Depth        int   // NOTE: largest depth of the document
	MaxDepth     *int  // NOTE: max depth to analyze
	NameLens     []int // NOTE: length of field names
	TotalObjects int64 // NOTE: total number of objects
	Stats        ObjectStats
}

// TypeStats stores statistics about a specific value type
type TypeStats struct {
	Type    string       `json:"Type"`
	Count   int64        `json:"Count"`
	Props   *ObjectStats `json:"Props,omitempty"`
	Items   *[]TypeStats `json:"Items,omitempty"`
	Subtype *string      `json:"Subtype,omitempty"`
}

func AnalyzeBSON(root *RootObject, elements []bson.RawElement, stats *ObjectStats) error {
	root.CurrDepth++
	defer func() {
		root.CurrDepth--
	}()

	if root.MaxDepth != nil && root.CurrDepth > *root.MaxDepth {
		return nil
	}

	if root.CurrDepth > root.Depth {
		root.Depth = root.CurrDepth
		root.NameLens = append(root.NameLens, 0)
	}

	for _, elm := range elements {
		key := elm.Key()
		t := elm.Value().Type.String()

		if _, ok := (*stats)[key]; !ok {
			(*stats)[key] = []TypeStats{}

			if len(key) > root.NameLens[root.CurrDepth-1] {
				root.NameLens[root.CurrDepth-1] = len(key)
			}
		}

		found := false
		for i, ts := range (*stats)[key] {
			if ts.Type == t {
				st := &(*stats)[key][i]
				newProps, err := handleEmbeddedDocument(root, elm, st.Props)
				if err != nil {
					return err
				}

				st.Count++
				st.Props = newProps
				found = true
				break
			}
		}
		if found {
			continue
		}

		var props *ObjectStats = nil
		// var items *ObjectStats = nil
		// var subType *ObjectStats = nil

		props, err := handleEmbeddedDocument(root, elm, nil)
		if err != nil {
			return err
		}

		(*stats)[key] = append((*stats)[key], TypeStats{
			Type:  t,
			Count: 1,
			Props: props,
		})
	}

	return nil
}

func handleEmbeddedDocument(root *RootObject, element bson.RawElement, props *ObjectStats) (*ObjectStats, error) {
	if element.Value().Type != bson.TypeEmbeddedDocument {
		return nil, nil
	}

	if props == nil {
		props = &ObjectStats{}
	}

	doc, ok := element.Value().DocumentOK()
	if !ok {
		return nil, ErrInvalidDocument
	}

	elements, err := doc.Elements()
	if err != nil {
		return nil, err
	}

	err = AnalyzeBSON(root, elements, props)
	if err != nil {
		return nil, err
	}

	return props, nil
}
