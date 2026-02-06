package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func OutputJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func MDHeader(level int, title string) {
	fmt.Printf("%s %s\n\n", strings.Repeat("#", level), title)
}

func MDTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}
	fmt.Printf("| %s |\n", strings.Join(headers, " | "))
	seps := make([]string, len(headers))
	for i := range seps {
		seps[i] = "---"
	}
	fmt.Printf("| %s |\n", strings.Join(seps, " | "))
	for _, row := range rows {
		// Pad row to match headers length
		for len(row) < len(headers) {
			row = append(row, "")
		}
		fmt.Printf("| %s |\n", strings.Join(row, " | "))
	}
	fmt.Println()
}
