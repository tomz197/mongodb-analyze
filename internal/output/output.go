package output

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/tomz197/mongodb-analyze/internal/analyze"
)

func PrintTable(root *analyze.RootObject) {
	/*
		Name | Type | Count | Percent
		------------------------------
		...
	*/

	fmt.Println()
	printHeader(root)

	printSeparator(root)

	root.CurrDepth = -1
	printRow(root, &root.Stats)

	printSeparator(root)

	fmt.Printf("Total objects: %d\n", root.TotalObjects)
	fmt.Printf("Max depth: %d\n", root.Depth)
	fmt.Printf("Name lengths: %v\n", root.NameLens)
	fmt.Println()
}

func printHeader(root *analyze.RootObject) {
	fillerAfter := ""
	for i := 1; i < root.Depth; i++ {
		fillerAfter += " | "
		for j := 0; j < int(root.NameLens[i]); j++ {
			fillerAfter += " "
		}
	}

	fmt.Printf(" %-*s%s | %-20s | %-10s | %-15s\n", root.NameLens[root.CurrDepth], "Name", fillerAfter, "Type", "Count", "Occurrence[%]")
}

func printSeparator(root *analyze.RootObject) {
	filler := ""
	for i := 0; i < root.Depth; i++ {
		filler += "---"
		for j := 0; j < int(root.NameLens[i]); j++ {
			filler += "-"
		}
	}
	fmt.Printf("%s------------------------------------------------------\n", filler)
}

func printRow(root *analyze.RootObject, stats *analyze.ObjectStats) {
	root.CurrDepth++
	defer func() {
		root.CurrDepth--
	}()
	if root.MaxDepth != nil && root.CurrDepth >= *root.MaxDepth {
		return
	}

	fillerBefore := " "
	for i := 0; i < int(root.CurrDepth); i++ {
		fillerBefore += ">"
		for j := 1; j < int(root.NameLens[i]); j++ {
			fillerBefore += " "
		}
		fillerBefore += " | "
	}

	fillerAfter := ""
	for i := root.CurrDepth + 1; i < root.Depth; i++ {
		fillerAfter += " | "
		for j := 0; j < int(root.NameLens[i]); j++ {
			fillerAfter += " "
		}
	}

	for _, kv := range getSorted(*stats) {
		for _, stat := range kv.Val {
			percent := float64(stat.Count) / float64(root.TotalObjects) * 100
			fmt.Printf("%s%-*s%s | %-20s | %-10d | %-15.2f\n",
				fillerBefore, root.NameLens[root.CurrDepth], kv.Key, fillerAfter, stat.Type, stat.Count, percent)

			if stat.Props != nil {
				printRow(root, stat.Props)
			}
		}
	}
}

type kv[T any] struct {
	Key string
	Val T
}

func getSorted[T any](m map[string]T) []kv[T] {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	res := make([]kv[T], len(keys))
	for i, k := range keys {
		res[i] = kv[T]{k, m[k]}
	}

	return res
}

func PrintJSON(anal *analyze.RootObject) {
	out, _ := json.MarshalIndent(anal.Stats, "", "  ")

	fmt.Println(string(out))
}
