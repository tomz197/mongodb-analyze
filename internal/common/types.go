package common

type RootObject struct {
	CurrDepth    int         // NOTE: current depth of the document
	Depth        int         // NOTE: largest depth of the document
	MaxDepth     *int        // NOTE: max depth to analyze
	MaxNameLen   int         // NOTE: length of the longest field name
	MaxTypeLen   int         // NOTE: length of the longest type name
	TotalObjects int64       // NOTE: total number of objects
	Stats        ObjectStats `json:"Stats"`
}

type ObjectStats map[string][]TypeStat

type TypeStat interface {
	// GetType returns the BSON type name (e.g., "string", "array", "binary").
	GetType() string
	// TypeDisplay returns a human-friendly type string, possibly including details
	// like binary subtypes or array composition (e.g., "array[string, int]").
	TypeDisplay() string
	// GetCount returns how many times this type has been observed for the field.
	GetCount() int64
	// Increment increases the observation count by 1.
	Increment()
	// GetProps returns nested field statistics (for embedded documents or arrays of documents).
	// Should return nil if the type does not have nested properties.
	GetProps() *ObjectStats
}

type BaseTypeStat struct {
	Type  string `json:"Type"`
	Count int64  `json:"Count"`
}

func (b *BaseTypeStat) GetType() string        { return b.Type }
func (b *BaseTypeStat) GetCount() int64        { return b.Count }
func (b *BaseTypeStat) Increment()             { b.Count++ }
func (b *BaseTypeStat) GetProps() *ObjectStats { return nil }
func (b *BaseTypeStat) TypeDisplay() string    { return b.Type }

type ScalarTypeStat struct {
	BaseTypeStat
}

func NewScalarTypeStat(typeName string) *ScalarTypeStat {
	return &ScalarTypeStat{BaseTypeStat: BaseTypeStat{Type: typeName, Count: 0}}
}

type DocumentTypeStat struct {
	BaseTypeStat
	Props ObjectStats `json:"Props"`
}

func NewDocumentTypeStat() *DocumentTypeStat {
	return &DocumentTypeStat{BaseTypeStat: BaseTypeStat{Type: "embedded document", Count: 0}, Props: ObjectStats{}}
}

func (d *DocumentTypeStat) GetProps() *ObjectStats { return &d.Props }
func (d *DocumentTypeStat) TypeDisplay() string    { return d.Type }

type ArrayTypeStat struct {
	BaseTypeStat
	Items     map[string]int `json:"Items"`
	ItemProps *ObjectStats   `json:"ItemProps,omitempty"`
}

func NewArrayTypeStat() *ArrayTypeStat {
	return &ArrayTypeStat{BaseTypeStat: BaseTypeStat{Type: "array", Count: 0}, Items: map[string]int{}, ItemProps: nil}
}

func (a *ArrayTypeStat) GetProps() *ObjectStats { return a.ItemProps }

func (a *ArrayTypeStat) TypeDisplay() string {
	if len(a.Items) == 0 {
		return a.Type
	}
	keys := make([]string, 0, len(a.Items))
	for k := range a.Items {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		j := i
		for j > 0 && keys[j-1] > keys[j] {
			keys[j-1], keys[j] = keys[j], keys[j-1]
			j--
		}
	}
	res := a.Type + "["
	for i, k := range keys {
		res += k
		if i < len(keys)-1 {
			res += ", "
		}
	}
	res += "]"
	return res
}

type BinaryTypeStat struct {
	BaseTypeStat
	Subtype     byte   `json:"Subtype"`
	SubtypeName string `json:"SubtypeName"`
}

func NewBinaryTypeStat(subtype byte, subtypeName string) *BinaryTypeStat {
	return &BinaryTypeStat{BaseTypeStat: BaseTypeStat{Type: "binary", Count: 0}, Subtype: subtype, SubtypeName: subtypeName}
}

func (b *BinaryTypeStat) TypeDisplay() string {
	if b.SubtypeName == "" {
		return b.Type
	}
	return b.Type + " - " + b.SubtypeName
}

// Ensuring that all type stats implement the TypeStat interface
var _ TypeStat = &BaseTypeStat{}
var _ TypeStat = &ScalarTypeStat{}
var _ TypeStat = &DocumentTypeStat{}
var _ TypeStat = &ArrayTypeStat{}
var _ TypeStat = &BinaryTypeStat{}
