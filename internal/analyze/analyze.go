package analyze

import (
	"errors"
	"fmt"
	"sort"

	"github.com/tomz197/mongodb-analyze/internal/common"
	"go.mongodb.org/mongo-driver/bson"
)

// Custom errors
var (
	ErrInvalidDocument = errors.New("invalid document")
)

func analyze(root *common.RootObject, elements []bson.RawElement, stats *common.ObjectStats) error {
	root.CurrDepth++
	defer func() {
		root.CurrDepth--
	}()

	if root.MaxDepth != nil && root.CurrDepth > *root.MaxDepth {
		return nil
	}

	if root.CurrDepth > root.Depth {
		root.Depth = root.CurrDepth
	}

	for _, elm := range elements {
		key := elm.Key()
		t := elm.Value().Type.String()

		t, err := handleBinarySubtype(elm)
		if err != nil {
			return ErrInvalidDocument
		}

		t, err = handleArray(elm)
		if err != nil {
			return ErrInvalidDocument
		}

		if _, ok := (*stats)[key]; !ok {
			(*stats)[key] = []common.TypeStats{}

			if len(key) > root.MaxNameLen {
				root.MaxNameLen = len(key)
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

		if len(t) > root.MaxTypeLen {
			root.MaxTypeLen = len(t)
		}

		var props *common.ObjectStats = nil
		// var items *ObjectStats = nil

		props, err = handleEmbeddedDocument(root, elm, nil)
		if err != nil {
			return err
		}

		(*stats)[key] = append((*stats)[key], common.TypeStats{
			Type:  t,
			Count: 1,
			Props: props,
		})
	}

	return nil
}

func handleEmbeddedDocument(root *common.RootObject, element bson.RawElement, props *common.ObjectStats) (*common.ObjectStats, error) {
	if element.Value().Type != bson.TypeEmbeddedDocument {
		return nil, nil
	}

	if props == nil {
		props = &common.ObjectStats{}
	}

	doc, ok := element.Value().DocumentOK()
	if !ok {
		return nil, ErrInvalidDocument
	}

	elements, err := doc.Elements()
	if err != nil {
		return nil, err
	}

	err = analyze(root, elements, props)
	if err != nil {
		return nil, err
	}

	return props, nil
}

var bsonBinarySubtypes = map[byte]string{
	0x00: "Generic binary subtype",
	0x01: "Function",
	0x02: "Binary (old)",
	0x03: "UUID (old)",
	0x04: "UUID",
	0x05: "MD5",
	0x06: "Encrypted BSON value",
	0x07: "Compressed time series data",
	0x08: "Sensitive data",
	0x09: "Vector data",
	0x80: "User defined custom data",
}

func handleBinarySubtype(element bson.RawElement) (string, error) {
	val := element.Value()
	if val.Type != bson.TypeBinary {
		return val.Type.String(), nil
	}

	subtype, _, ok := val.BinaryOK()
	if !ok {
		return "", ErrInvalidDocument
	}

	subtypeStr, ok := bsonBinarySubtypes[subtype]
	if !ok {
		return "Unknown", nil
	}

	return val.Type.String() + " - " + subtypeStr, nil
}

func handleArray(arr bson.RawElement) (string, error) {
	val := arr.Value()
	if val.Type != bson.TypeArray {
		return val.Type.String(), nil
	}

	arrRaw, ok := val.ArrayOK()
	if !ok {
		return "", ErrInvalidDocument
	}

	elements, err := arrRaw.Elements()
	if err != nil {
		return "", err
	}

	types := make(map[string]int)
	for _, elm := range elements {
		t := elm.Value().Type.String()
		types[t]++
	}

	var keys []string
	for k := range types {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var res string
	for i, k := range keys {
		res += k
		if i < len(keys)-1 {
			res += ", "
		}
	}

	return fmt.Sprintf("%s[%s]", arr.Value().Type.String(), res), nil
}
