package cmd

import (
	"fmt"
	"time"

	"github.com/dkd/ccs/internal/display"
	"github.com/dkd/ccs/internal/store"
)

func Period(period string) error {
	stats, err := store.LoadStatsCache()
	if err != nil {
		return fmt.Errorf("loading stats cache: %w", err)
	}

	now := time.Now()
	var startDate time.Time
	var title string

	switch period {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		title = "Today"
	case "week":
		// Go back to Monday of current week
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		startDate = time.Date(now.Year(), now.Month(), now.Day()-int(weekday)+1, 0, 0, 0, 0, now.Location())
		title = "This Week"
	case "month":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		title = "This Month"
	}

	startStr := startDate.Format("2006-01-02")

	// Filter daily activity
	var totalMessages, totalSessions, totalToolCalls int
	var days []store.DailyActivity
	for _, d := range stats.DailyActivity {
		if d.Date >= startStr {
			days = append(days, d)
			totalMessages += d.MessageCount
			totalSessions += d.SessionCount
			totalToolCalls += d.ToolCallCount
		}
	}

	// Filter tokens
	var tokensByModel = make(map[string]int)
	for _, dt := range stats.DailyModelTokens {
		if dt.Date >= startStr {
			for model, tokens := range dt.TokensByModel {
				tokensByModel[model] += tokens
			}
		}
	}

	// Load sessions for this period
	sessions, sessErr := store.ListSessionsAfter(startDate)

	if display.IsJSON() {
		return periodJSON(period, startStr, totalSessions, totalMessages, totalToolCalls, tokensByModel, days, sessions)
	}
	if display.IsMD() {
		return periodMD(title, startStr, totalSessions, totalMessages, totalToolCalls, tokensByModel, days, sessions)
	}

	fmt.Println(display.BoldCyan(title))
	fmt.Printf("Since %s\n\n", startStr)

	display.Box("Activity", func() {
		fmt.Printf("  Sessions    %s\n", display.Bold(display.FormatNumber(totalSessions)))
		fmt.Printf("  Messages    %s\n", display.Bold(display.FormatNumber(totalMessages)))
		fmt.Printf("  Tool calls  %s\n", display.Bold(display.FormatNumber(totalToolCalls)))
		if totalSessions > 0 {
			avg := totalMessages / totalSessions
			fmt.Printf("  Avg/session %s msgs\n", display.Bold(display.FormatNumber(avg)))
		}
	})
	fmt.Println()

	if len(tokensByModel) > 0 {
		display.Box("Tokens (output)", func() {
			for model, tokens := range tokensByModel {
				fmt.Printf("  %-14s  %s\n",
					display.ModelShort(model),
					display.Bold(display.FormatTokens(tokens)))
			}
		})
		fmt.Println()
	}

	// Daily breakdown
	if len(days) > 1 {
		display.Box("Daily Breakdown", func() {
			maxMsg := 0
			for _, d := range days {
				if d.MessageCount > maxMsg {
					maxMsg = d.MessageCount
				}
			}
			for _, d := range days {
				bar := display.Bar(d.MessageCount, maxMsg, 20)
				fmt.Printf("  %s  %s %s msgs, %d sessions\n",
					d.Date,
					bar,
					display.FormatNumber(d.MessageCount),
					d.SessionCount)
			}
		})
		fmt.Println()
	}

	if sessErr == nil && len(sessions) > 0 {
		display.Box("Sessions", func() {
			limit := 15
			if len(sessions) < limit {
				limit = len(sessions)
			}
			for i := 0; i < limit; i++ {
				s := sessions[i]
				prompt := display.Truncate(s.FirstPrompt, 50)
				if prompt == "" {
					prompt = display.Dim("(no prompt)")
				}
				created := ""
				if t, err := time.Parse(time.RFC3339, s.Created); err == nil {
					created = display.RelativeTime(t)
				}
				fmt.Printf("  %s  %3d msgs  %-10s  %s\n",
					display.Dim(s.SessionID[:8]),
					s.MessageCount,
					created,
					prompt)
			}
			if len(sessions) > limit {
				fmt.Printf("  %s\n", display.Dim(fmt.Sprintf("... and %d more", len(sessions)-limit)))
			}
		})
		fmt.Println()
	}

	return nil
}

func periodJSON(period, since string, totalSessions, totalMessages, totalToolCalls int, tokensByModel map[string]int, days []store.DailyActivity, sessions []store.SessionEntry) error {
	shortTokens := make(map[string]int)
	for model, tokens := range tokensByModel {
		shortTokens[display.ModelShort(model)] = tokens
	}

	type jsonDay struct {
		Date     string `json:"date"`
		Messages int    `json:"messages"`
		Sessions int    `json:"sessions"`
		Tools    int    `json:"toolCalls"`
	}
	var jsonDays []jsonDay
	for _, d := range days {
		jsonDays = append(jsonDays, jsonDay{
			Date:     d.Date,
			Messages: d.MessageCount,
			Sessions: d.SessionCount,
			Tools:    d.ToolCallCount,
		})
	}

	type jsonSession struct {
		SessionID   string `json:"sessionId"`
		Messages    int    `json:"messages"`
		Created     string `json:"created"`
		FirstPrompt string `json:"firstPrompt"`
	}
	var jsonSessions []jsonSession
	for _, s := range sessions {
		jsonSessions = append(jsonSessions, jsonSession{
			SessionID:   s.SessionID,
			Messages:    s.MessageCount,
			Created:     s.Created,
			FirstPrompt: s.FirstPrompt,
		})
	}

	data := map[string]any{
		"period": period,
		"since":  since,
		"activity": map[string]int{
			"sessions":  totalSessions,
			"messages":  totalMessages,
			"toolCalls": totalToolCalls,
		},
		"tokens":   shortTokens,
		"days":     jsonDays,
		"sessions": jsonSessions,
	}
	return OutputJSON(data)
}

func periodMD(title, since string, totalSessions, totalMessages, totalToolCalls int, tokensByModel map[string]int, days []store.DailyActivity, sessions []store.SessionEntry) error {
	MDHeader(2, title)
	fmt.Printf("Since %s\n\n", since)

	MDHeader(3, "Activity")
	fmt.Printf("- **Sessions:** %s\n", display.FormatNumber(totalSessions))
	fmt.Printf("- **Messages:** %s\n", display.FormatNumber(totalMessages))
	fmt.Printf("- **Tool calls:** %s\n", display.FormatNumber(totalToolCalls))
	if totalSessions > 0 {
		avg := totalMessages / totalSessions
		fmt.Printf("- **Avg/session:** %s msgs\n", display.FormatNumber(avg))
	}
	fmt.Println()

	if len(tokensByModel) > 0 {
		MDHeader(3, "Tokens (output)")
		tHeaders := []string{"Model", "Tokens"}
		var tRows [][]string
		for model, tokens := range tokensByModel {
			tRows = append(tRows, []string{
				display.ModelShort(model),
				display.FormatTokens(tokens),
			})
		}
		MDTable(tHeaders, tRows)
	}

	if len(days) > 1 {
		MDHeader(3, "Daily Breakdown")
		dHeaders := []string{"Date", "Messages", "Sessions"}
		var dRows [][]string
		for _, d := range days {
			dRows = append(dRows, []string{
				d.Date,
				display.FormatNumber(d.MessageCount),
				fmt.Sprintf("%d", d.SessionCount),
			})
		}
		MDTable(dHeaders, dRows)
	}

	if len(sessions) > 0 {
		MDHeader(3, "Sessions")
		sHeaders := []string{"ID", "Messages", "Created", "Prompt"}
		var sRows [][]string
		limit := 15
		if len(sessions) < limit {
			limit = len(sessions)
		}
		for i := 0; i < limit; i++ {
			s := sessions[i]
			prompt := display.Truncate(s.FirstPrompt, 50)
			if prompt == "" {
				prompt = "(no prompt)"
			}
			created := ""
			if t, err := time.Parse(time.RFC3339, s.Created); err == nil {
				created = display.RelativeTime(t)
			}
			sRows = append(sRows, []string{
				s.SessionID[:8],
				fmt.Sprintf("%d", s.MessageCount),
				created,
				prompt,
			})
		}
		MDTable(sHeaders, sRows)
		if len(sessions) > limit {
			fmt.Printf("... and %d more\n\n", len(sessions)-limit)
		}
	}

	return nil
}
