package cmd

import (
	"fmt"

	"github.com/dkd/ccs/internal/display"
	"github.com/dkd/ccs/internal/store"
)

func Refresh() error {
	fmt.Println(display.BoldCyan("Refreshing stats cache..."))

	stats, err := store.ComputeStats(func(done, total int) {
		fmt.Printf("\r  Scanning... %d/%d sessions", done, total)
	})
	if err != nil {
		return fmt.Errorf("computing stats: %w", err)
	}
	fmt.Println()

	if err := store.SaveStatsCache(stats); err != nil {
		return fmt.Errorf("saving stats: %w", err)
	}

	fmt.Printf("\n  %s %s sessions, %s messages\n",
		display.Green("Done."),
		display.Bold(display.FormatNumber(stats.TotalSessions)),
		display.Bold(display.FormatNumber(stats.TotalMessages)))
	fmt.Printf("  Cache updated for %s\n\n", display.Bold(stats.LastComputedDate))

	return nil
}
