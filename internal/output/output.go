package output

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/tomz197/mongodb-analyze/internal/common"
)

func PrintTable(root *common.RootObject) {
	/*
		Name | ...(subobjects)... | Type | Count | Occurrence[%]
		----------------------------------------------------------
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

func printHeader(root *common.RootObject) {
	fillerAfter := ""
	for i := 1; i < root.Depth; i++ {
		fillerAfter += " | "
		for j := 0; j < int(root.NameLens[i]); j++ {
			fillerAfter += " "
		}
	}

	fmt.Printf(" %-*s%s | %-*s | %-10s | %-15s\n", root.NameLens[root.CurrDepth], "Name", fillerAfter, root.MaxTypeLen, "Type", "Count", "Occurrence[%]")
}

func printSeparator(root *common.RootObject) {
	filler := ""
	for i := 0; i < root.Depth; i++ {
		filler += "---"
		for j := 0; j < int(root.NameLens[i]); j++ {
			filler += "-"
		}
	}
	for j := 0; j < root.MaxTypeLen; j++ {
		filler += "-"
	}
	fmt.Printf("%s----------------------------------\n", filler)
}

func printRow(root *common.RootObject, stats *common.ObjectStats) {
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
			fmt.Printf("%s%-*s%s | %-*s | %-10d | %-15.2f\n",
				fillerBefore, root.NameLens[root.CurrDepth], kv.Key, fillerAfter, root.MaxTypeLen, stat.Type, stat.Count, percent)

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

func PrintJSON(anal *common.RootObject) {
	out, _ := json.MarshalIndent(anal.Stats, "", "  ")

	fmt.Println(string(out))
}

func GetPrintProgress(total int64) (func(int64), func(int64)) {
	percOfDocs := total / 100

	printProgress := func(processed int64) {
		bar := ""
		for range processed / (percOfDocs * 5) {
			bar += "="
		}
		if len(bar) < 20 {
			bar += ">"
		}
		fmt.Printf("\r Progress [%-20s] %d%% (%d/%d)", bar, processed/percOfDocs, processed, total)
	}

	return func(processed int64) {
			if processed%percOfDocs != 0 {
				return
			}
			printProgress(processed)
		}, func(processed int64) {
			printProgress(processed)
			fmt.Println()
			fmt.Println("Done")
		}
}
