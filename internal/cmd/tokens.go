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

	if display.IsJSON() {
		return tokensJSON(stats)
	}
	if display.IsMD() {
		return tokensMD(stats)
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
			fmt.Printf("  %s\n", display.Bold(display.ModelShort(m.name)))
			fmt.Printf("    Input tokens     %s\n", display.FormatTokens(m.usage.InputTokens))
			fmt.Printf("    Output tokens    %s\n", display.Bold(display.FormatTokens(m.usage.OutputTokens)))
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

func tokensJSON(stats *store.StatsCache) error {
	models := make(map[string]any)
	for name, m := range stats.ModelUsage {
		models[display.ModelShort(name)] = map[string]int{
			"inputTokens":    m.InputTokens,
			"outputTokens":   m.OutputTokens,
			"cacheRead":      m.CacheReadInputTokens,
			"cacheCreation":  m.CacheCreationInputTokens,
		}
	}

	type dailyEntry struct {
		Date   string         `json:"date"`
		Tokens map[string]int `json:"tokens"`
	}
	days := stats.DailyModelTokens
	start := 0
	if len(days) > 14 {
		start = len(days) - 14
	}
	recent := days[start:]
	var dailyTokens []dailyEntry
	for _, d := range recent {
		shortTokens := make(map[string]int)
		for model, tokens := range d.TokensByModel {
			shortTokens[display.ModelShort(model)] = tokens
		}
		dailyTokens = append(dailyTokens, dailyEntry{
			Date:   d.Date,
			Tokens: shortTokens,
		})
	}

	data := map[string]any{
		"models":      models,
		"dailyTokens": dailyTokens,
	}
	return OutputJSON(data)
}

func tokensMD(stats *store.StatsCache) error {
	MDHeader(2, "Token Usage")

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

	MDHeader(3, "By Model")
	headers := []string{"Model", "Input", "Output", "Cache Read", "Cache Creation"}
	var rows [][]string
	for _, m := range models {
		rows = append(rows, []string{
			display.ModelShort(m.name),
			display.FormatTokens(m.usage.InputTokens),
			display.FormatTokens(m.usage.OutputTokens),
			display.FormatTokens(m.usage.CacheReadInputTokens),
			display.FormatTokens(m.usage.CacheCreationInputTokens),
		})
	}
	MDTable(headers, rows)

	MDHeader(3, "Daily Output Tokens (last 14 days)")
	days := stats.DailyModelTokens
	start := 0
	if len(days) > 14 {
		start = len(days) - 14
	}
	recent := days[start:]

	dHeaders := []string{"Date", "Output Tokens"}
	var dRows [][]string
	for _, d := range recent {
		total := 0
		for _, t := range d.TokensByModel {
			total += t
		}
		dRows = append(dRows, []string{d.Date, display.FormatTokens(total)})
	}
	MDTable(dHeaders, dRows)

	return nil
}
