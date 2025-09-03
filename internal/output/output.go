package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/tomz197/mongodb-analyze/internal/common"
)

func PrintTable(root *common.RootObject, out *os.File) {
	/*
		Name | Type | Count | Occurrence[%]
		-------------------------------------
		...
	*/

	fmt.Fprintln(out, "MongoDB Document Analysis")
	printHeader(root, out)

	printSeparator(root, out)

	root.CurrDepth = -1
	root.MaxNameLen += 3 * (root.Depth)
	printRow(root, &root.Stats, out)

	printSeparator(root, out)

	fmt.Fprintf(out, "Total objects: %d\n", root.TotalObjects)
	fmt.Fprintf(out, "Max depth: %d\n", root.Depth)
	fmt.Fprintln(out)
}

func printHeader(root *common.RootObject, out *os.File) {
	fmt.Fprintf(out, " %-*s | %-*s | %-10s | %-15s\n", root.MaxNameLen+3, "Name", root.MaxTypeLen, "Type", "Count", "Occurrence[%]")
}

func printSeparator(root *common.RootObject, out *os.File) {
	filler := ""
	for i := 0; i < root.Depth; i++ {
		filler += "---"
	}
	for j := 0; j < int(root.MaxNameLen); j++ {
		filler += "-"
	}
	for j := 0; j < root.MaxTypeLen; j++ {
		filler += "-"
	}
	fmt.Fprintf(out, "%s----------------------------------\n", filler)
}

func printRow(root *common.RootObject, stats *common.ObjectStats, out *os.File) {
	root.CurrDepth++
	defer func() {
		root.CurrDepth--
	}()
	if root.MaxDepth != nil && root.CurrDepth >= *root.MaxDepth {
		return
	}

	fillerBefore := " "
	for i := 0; i < int(root.CurrDepth); i++ {
		fillerBefore += " > "
	}

	for _, kv := range getSorted(*stats) {
		for _, stat := range kv.Val {
			percent := float64(stat.GetCount()) / float64(root.TotalObjects) * 100
			fmt.Fprintf(out, "%s%-*s | %-*s | %-10d | %-15.2f\n",
				fillerBefore, root.MaxNameLen-(3*(root.CurrDepth+1)), kv.Key, root.MaxTypeLen, stat.TypeDisplay(), stat.GetCount(), percent)

			if props := stat.GetProps(); props != nil {
				printRow(root, props, out)
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

func PrintJSON(anal *common.RootObject, out *os.File) {
	marshalled, _ := json.MarshalIndent(anal.Stats, "", "  ")

	fmt.Fprintln(out, string(marshalled))
}

func GetPrintProgress(total int64) (func(int64), func(int64)) {
	percOfDocs := total / 100
	if percOfDocs == 0 {
		percOfDocs = 1
	}

	printProgress := func(processed int64) {
		bar := ""
		for range processed / (percOfDocs * 5) {
			bar += "="
		}
		if len(bar) < 20 {
			bar += ">"
		}
		fmt.Fprintf(os.Stdout, "\r Progress [%-20s] %d%% (%d/%d)", bar, processed/percOfDocs, processed, total)
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
