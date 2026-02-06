package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/dkd/ccs/internal/display"
	"github.com/dkd/ccs/internal/store"
)

func All() error {
	if display.IsJSON() {
		return allJSON()
	}

	// MD and terminal: run sub-commands sequentially
	if err := Summary(); err != nil {
		return err
	}

	fmt.Println()
	if err := Projects(); err != nil {
		return err
	}

	fmt.Println()
	if err := Sessions(nil); err != nil {
		return err
	}

	fmt.Println()
	if err := Tokens(); err != nil {
		return err
	}

	return nil
}

func allJSON() error {
	stats, err := store.LoadStatsCache()
	if err != nil {
		return fmt.Errorf("loading stats cache: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	todayStats := findDay(stats.DailyActivity, today)

	// Summary
	var totalIn, totalOut, totalCache int
	for _, m := range stats.ModelUsage {
		totalIn += m.InputTokens
		totalOut += m.OutputTokens
		totalCache += m.CacheReadInputTokens
	}

	models := make(map[string]any)
	for name, m := range stats.ModelUsage {
		models[display.ModelShort(name)] = map[string]int{
			"inputTokens":  m.InputTokens,
			"outputTokens": m.OutputTokens,
			"cacheRead":    m.CacheReadInputTokens,
			"cacheCreation": m.CacheCreationInputTokens,
		}
	}

	type hourEntry struct {
		Hour  int `json:"hour"`
		Count int `json:"sessions"`
	}
	var hours []hourEntry
	for h, c := range stats.HourCounts {
		var hour int
		fmt.Sscanf(h, "%d", &hour)
		hours = append(hours, hourEntry{hour, c})
	}
	sort.Slice(hours, func(i, j int) bool {
		return hours[i].Count > hours[j].Count
	})

	todayData := map[string]any{}
	if todayStats != nil {
		todayData = map[string]any{
			"sessions":  todayStats.SessionCount,
			"messages":  todayStats.MessageCount,
			"toolCalls": todayStats.ToolCallCount,
		}
	}

	summary := map[string]any{
		"overview": map[string]any{
			"sessions":         stats.TotalSessions,
			"messages":         stats.TotalMessages,
			"outputTokens":    totalOut,
			"inputTokens":     totalIn,
			"cacheReadTokens": totalCache,
			"trackingSince":   stats.FirstSessionDate,
			"lastComputed":    stats.LastComputedDate,
		},
		"today":     todayData,
		"models":    models,
		"peakHours": hours,
	}
	if stats.LongestSession.SessionID != "" {
		summary["longestSession"] = map[string]any{
			"sessionId": stats.LongestSession.SessionID,
			"duration":  display.FormatDuration(stats.LongestSession.Duration),
			"messages":  stats.LongestSession.MessageCount,
		}
	}

	// Projects
	allProjects, _ := store.LoadAllProjects()
	sort.Slice(allProjects, func(i, j int) bool {
		return allProjects[i].MessageCount > allProjects[j].MessageCount
	})
	type jsonProject struct {
		Path       string `json:"path"`
		Sessions   int    `json:"sessions"`
		Messages   int    `json:"messages"`
		LastActive string `json:"lastActive"`
	}
	var projectsOut []jsonProject
	for _, p := range allProjects {
		name := p.Path
		if name == "" {
			name = p.DirName
		}
		la := ""
		if !p.LastActive.IsZero() {
			la = p.LastActive.Format(time.RFC3339)
		}
		projectsOut = append(projectsOut, jsonProject{
			Path:       name,
			Sessions:   p.SessionCount,
			Messages:   p.MessageCount,
			LastActive: la,
		})
	}

	// Sessions
	allSessions, _ := store.ListAllSessions("")
	type jsonSession struct {
		SessionID   string `json:"sessionId"`
		Messages    int    `json:"messages"`
		Created     string `json:"created"`
		FirstPrompt string `json:"firstPrompt"`
		Branch      string `json:"branch,omitempty"`
		Sidechain   bool   `json:"sidechain,omitempty"`
	}
	limit := 20
	if limit > len(allSessions) {
		limit = len(allSessions)
	}
	var sessionsOut []jsonSession
	for i := 0; i < limit; i++ {
		s := allSessions[i]
		sessionsOut = append(sessionsOut, jsonSession{
			SessionID:   s.SessionID,
			Messages:    s.MessageCount,
			Created:     s.Created,
			FirstPrompt: s.FirstPrompt,
			Branch:      s.GitBranch,
			Sidechain:   s.IsSidechain,
		})
	}

	// Tokens
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
		"summary":  summary,
		"projects": projectsOut,
		"sessions": sessionsOut,
		"tokens": map[string]any{
			"models":      models,
			"dailyTokens": dailyTokens,
		},
	}

	return OutputJSON(data)
}
