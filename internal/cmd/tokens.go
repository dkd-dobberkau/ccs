package cmd

import (
	"fmt"
	"sort"

	"github.com/dkd/ccs/internal/display"
	"github.com/dkd/ccs/internal/store"
)

func Tokens() error {
	stats, err := store.LoadStatsCache()
	if err != nil {
		return fmt.Errorf("loading stats cache: %w", err)
	}

	fmt.Println(display.BoldCyan("Token Usage"))
	fmt.Println()

	// Per-model breakdown
	type modelEntry struct {
		name  string
		usage store.ModelUsage
	}
	var models []modelEntry
	for name, usage := range stats.ModelUsage {
		models = append(models, modelEntry{name, usage})
	}
	sort.Slice(models, func(i, j int) bool {
		return models[i].usage.OutputTokens > models[j].usage.OutputTokens
	})

	display.Box("By Model", func() {
		for _, m := range models {
			fmt.Printf("  %s\n", display.BoldWhite(display.ModelShort(m.name)))
			fmt.Printf("    Input tokens     %s\n", display.FormatTokens(m.usage.InputTokens))
			fmt.Printf("    Output tokens    %s\n", display.BoldWhite(display.FormatTokens(m.usage.OutputTokens)))
			fmt.Printf("    Cache read       %s\n", display.FormatTokens(m.usage.CacheReadInputTokens))
			fmt.Printf("    Cache creation   %s\n", display.FormatTokens(m.usage.CacheCreationInputTokens))
			fmt.Println()
		}
	})
	fmt.Println()

	// Daily output tokens (last 14 days)
	display.Box("Daily Output Tokens (last 14 days)", func() {
		days := stats.DailyModelTokens
		start := 0
		if len(days) > 14 {
			start = len(days) - 14
		}
		recent := days[start:]

		// Find max for bar
		maxTokens := 0
		for _, d := range recent {
			total := 0
			for _, t := range d.TokensByModel {
				total += t
			}
			if total > maxTokens {
				maxTokens = total
			}
		}

		for _, d := range recent {
			total := 0
			for _, t := range d.TokensByModel {
				total += t
			}
			bar := display.Bar(total, maxTokens, 20)
			fmt.Printf("  %s  %s %s\n", d.Date, bar, display.FormatTokens(total))
		}
	})
	fmt.Println()

	return nil
}
