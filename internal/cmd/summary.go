package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/dkd/ccs/internal/display"
	"github.com/dkd/ccs/internal/store"
)

func Summary() error {
	stats, err := store.LoadStatsCache()
	if err != nil {
		return fmt.Errorf("loading stats cache: %w", err)
	}

	// Header
	firstDate := stats.FirstSessionDate
	if t, err := time.Parse(time.RFC3339, firstDate); err == nil {
		firstDate = t.Format("2006-01-02")
	}
	fmt.Println(display.BoldCyan("Claude Code Summary"))
	fmt.Printf("Tracking since %s (last updated: %s)\n\n", firstDate, stats.LastComputedDate)

	// Staleness warning
	today := time.Now().Format("2006-01-02")
	if stats.LastComputedDate != today {
		fmt.Printf("%s Stats last computed on %s (today: %s)\n\n",
			display.Yellow("⚠"),
			stats.LastComputedDate,
			today)
	}

	// Overall stats
	display.Box("Overview", func() {
		fmt.Printf("  Sessions    %s\n", display.BoldWhite(display.FormatNumber(stats.TotalSessions)))
		fmt.Printf("  Messages    %s\n", display.BoldWhite(display.FormatNumber(stats.TotalMessages)))

		// Total tokens across all models
		var totalIn, totalOut, totalCache int
		for _, m := range stats.ModelUsage {
			totalIn += m.InputTokens
			totalOut += m.OutputTokens
			totalCache += m.CacheReadInputTokens
		}
		fmt.Printf("  Tokens out  %s\n", display.BoldWhite(display.FormatTokens(totalOut)))
		fmt.Printf("  Cache read  %s\n", display.BoldWhite(display.FormatTokens(totalCache)))
	})
	fmt.Println()

	// Today's stats
	todayStats := findDay(stats.DailyActivity, today)
	display.Box("Today", func() {
		if todayStats != nil {
			fmt.Printf("  Sessions    %s\n", display.BoldWhite(display.FormatNumber(todayStats.SessionCount)))
			fmt.Printf("  Messages    %s\n", display.BoldWhite(display.FormatNumber(todayStats.MessageCount)))
			fmt.Printf("  Tool calls  %s\n", display.BoldWhite(display.FormatNumber(todayStats.ToolCallCount)))
		} else {
			fmt.Printf("  %s\n", display.Dim("No activity recorded for today"))
		}
	})
	fmt.Println()

	// Model usage
	display.Box("Models", func() {
		type modelEntry struct {
			name   string
			usage  store.ModelUsage
		}
		var models []modelEntry
		for name, usage := range stats.ModelUsage {
			models = append(models, modelEntry{name, usage})
		}
		sort.Slice(models, func(i, j int) bool {
			return models[i].usage.OutputTokens > models[j].usage.OutputTokens
		})
		for _, m := range models {
			fmt.Printf("  %-14s  out: %-10s  cache: %s\n",
				display.BoldWhite(display.ModelShort(m.name)),
				display.FormatTokens(m.usage.OutputTokens),
				display.FormatTokens(m.usage.CacheReadInputTokens))
		}
	})
	fmt.Println()

	// Peak hours
	display.Box("Peak Hours", func() {
		type hourEntry struct {
			hour  int
			count int
		}
		var hours []hourEntry
		for h, c := range stats.HourCounts {
			var hour int
			fmt.Sscanf(h, "%d", &hour)
			hours = append(hours, hourEntry{hour, c})
		}
		sort.Slice(hours, func(i, j int) bool {
			return hours[i].count > hours[j].count
		})
		// Show top 5
		limit := 5
		if len(hours) < limit {
			limit = len(hours)
		}
		for i := 0; i < limit; i++ {
			h := hours[i]
			bar := display.Bar(h.count, hours[0].count, 20)
			fmt.Printf("  %02d:00  %s %d sessions\n", h.hour, bar, h.count)
		}
	})
	fmt.Println()

	// Longest session
	if stats.LongestSession.SessionID != "" {
		display.Box("Longest Session", func() {
			fmt.Printf("  ID        %s\n", display.Dim(stats.LongestSession.SessionID[:8]))
			fmt.Printf("  Duration  %s\n", display.BoldWhite(display.FormatDuration(stats.LongestSession.Duration)))
			fmt.Printf("  Messages  %s\n", display.BoldWhite(display.FormatNumber(stats.LongestSession.MessageCount)))
		})
		fmt.Println()
	}

	return nil
}

func findDay(days []store.DailyActivity, date string) *store.DailyActivity {
	for i := range days {
		if days[i].Date == date {
			return &days[i]
		}
	}
	return nil
}
