package display

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// Table provides a simple tabwriter wrapper
type Table struct {
	w *tabwriter.Writer
}

// NewTable creates a table that writes to stdout
func NewTable() *Table {
	return &Table{
		w: tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0),
	}
}

// Row writes a tab-separated row
func (t *Table) Row(cells ...string) {
	for i, c := range cells {
		if i > 0 {
			fmt.Fprint(t.w, "\t")
		}
		fmt.Fprint(t.w, c)
	}
	fmt.Fprintln(t.w)
}

// Flush flushes the tabwriter
func (t *Table) Flush() {
	t.w.Flush()
}
