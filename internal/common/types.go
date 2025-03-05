package common

// ObjectStats maps field names to their type statistics
type ObjectStats map[string][]TypeStats

// RootObject is has more statistics/flags
type RootObject struct {
	CurrDepth    int   // NOTE: current depth of the document
	Depth        int   // NOTE: largest depth of the document
	MaxDepth     *int  // NOTE: max depth to analyze
	NameLens     []int // NOTE: length of field names
	MaxTypeLen   int   // NOTE: length of the longest type name
	TotalObjects int64 // NOTE: total number of objects
	Stats        ObjectStats
}

// TypeStats stores statistics about a specific value type
type TypeStats struct {
	Type    string       `json:"Type"`
	Count   int64        `json:"Count"`
	Props   *ObjectStats `json:"Props,omitempty"`
	Subtype *string      `json:"Subtype,omitempty"`
}
