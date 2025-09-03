package analyze

import (
	"errors"

	"github.com/tomz197/mongodb-analyze/internal/common"
	"go.mongodb.org/mongo-driver/bson"
)

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
		val := elm.Value()

		if err := updateTypeStatsForField(root, key, val, stats); err != nil {
			return err
		}
	}

	return nil
}

func updateTypeStatsForField(root *common.RootObject, key string, val bson.RawValue, stats *common.ObjectStats) error {
	if _, ok := (*stats)[key]; !ok {
		(*stats)[key] = []common.TypeStat{}
		if len(key) > root.MaxNameLen {
			root.MaxNameLen = len(key)
		}
	}

	switch val.Type {
	case bson.TypeEmbeddedDocument:
		var docStat *common.DocumentTypeStat
		for _, ts := range (*stats)[key] {
			if v, ok := ts.(*common.DocumentTypeStat); ok {
				docStat = v
				break
			}
		}
		if docStat == nil {
			docStat = common.NewDocumentTypeStat()
			(*stats)[key] = append((*stats)[key], docStat)
		}
		docStat.Increment()
		newProps, err := descendIntoDocument(root, val, &docStat.Props)
		if err != nil {
			return err
		}
		docStat.Props = *newProps
		if l := len(docStat.TypeDisplay()); l > root.MaxTypeLen {
			root.MaxTypeLen = l
		}

	case bson.TypeArray:
		var arrStat *common.ArrayTypeStat
		for _, ts := range (*stats)[key] {
			if v, ok := ts.(*common.ArrayTypeStat); ok {
				arrStat = v
				break
			}
		}
		if arrStat == nil {
			arrStat = common.NewArrayTypeStat()
			(*stats)[key] = append((*stats)[key], arrStat)
		}
		arrStat.Increment()
		arr, ok := val.ArrayOK()
		if !ok {
			return ErrInvalidDocument
		}
		elements, err := arr.Elements()
		if err != nil {
			return err
		}
		for _, a := range elements {
			av := a.Value()
			name := elementTypeDisplay(av)
			arrStat.Items[name] = arrStat.Items[name] + 1
			if av.Type == bson.TypeEmbeddedDocument {
				if arrStat.ItemProps == nil {
					mp := common.ObjectStats{}
					arrStat.ItemProps = &mp
				}
				newProps, err := descendIntoDocument(root, av, arrStat.ItemProps)
				if err != nil {
					return err
				}
				*arrStat.ItemProps = *newProps
			}
		}
		if l := len(arrStat.TypeDisplay()); l > root.MaxTypeLen {
			root.MaxTypeLen = l
		}

	case bson.TypeBinary:
		subtype, _, ok := val.BinaryOK()
		if !ok {
			return ErrInvalidDocument
		}
		var binStat *common.BinaryTypeStat
		for _, ts := range (*stats)[key] {
			if v, ok := ts.(*common.BinaryTypeStat); ok {
				if v.Subtype == subtype {
					binStat = v
					break
				}
			}
		}
		if binStat == nil {
			subtypeName := lookupBinarySubtypeName(subtype)
			binStat = common.NewBinaryTypeStat(subtype, subtypeName)
			(*stats)[key] = append((*stats)[key], binStat)
		}
		binStat.Increment()
		if l := len(binStat.TypeDisplay()); l > root.MaxTypeLen {
			root.MaxTypeLen = l
		}

	default:
		typeName := val.Type.String()
		var sc *common.ScalarTypeStat
		for _, ts := range (*stats)[key] {
			if v, ok := ts.(*common.ScalarTypeStat); ok {
				if v.GetType() == typeName {
					sc = v
					break
				}
			}
		}
		if sc == nil {
			sc = common.NewScalarTypeStat(typeName)
			(*stats)[key] = append((*stats)[key], sc)
		}
		sc.Increment()
		if l := len(sc.TypeDisplay()); l > root.MaxTypeLen {
			root.MaxTypeLen = l
		}
	}

	return nil
}

func descendIntoDocument(root *common.RootObject, value bson.RawValue, props *common.ObjectStats) (*common.ObjectStats, error) {
	if value.Type != bson.TypeEmbeddedDocument {
		return nil, nil
	}

	if props == nil {
		mp := common.ObjectStats{}
		props = &mp
	}

	doc, ok := value.DocumentOK()
	if !ok {
		return nil, ErrInvalidDocument
	}

	elements, err := doc.Elements()
	if err != nil {
		return nil, err
	}

	if err := analyze(root, elements, props); err != nil {
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

func lookupBinarySubtypeName(subtype byte) string {
	if s, ok := bsonBinarySubtypes[subtype]; ok {
		return s
	}
	return "Unknown"
}

func elementTypeDisplay(v bson.RawValue) string {
	if v.Type == bson.TypeBinary {
		subtype, _, ok := v.BinaryOK()
		if !ok {
			return v.Type.String()
		}
		return v.Type.String() + " - " + lookupBinarySubtypeName(subtype)
	}
	return v.Type.String()
}
