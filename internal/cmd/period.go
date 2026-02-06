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

	fmt.Println(display.BoldCyan(title))
	fmt.Printf("Since %s\n\n", startStr)

	display.Box("Activity", func() {
		fmt.Printf("  Sessions    %s\n", display.BoldWhite(display.FormatNumber(totalSessions)))
		fmt.Printf("  Messages    %s\n", display.BoldWhite(display.FormatNumber(totalMessages)))
		fmt.Printf("  Tool calls  %s\n", display.BoldWhite(display.FormatNumber(totalToolCalls)))
		if totalSessions > 0 {
			avg := totalMessages / totalSessions
			fmt.Printf("  Avg/session %s msgs\n", display.BoldWhite(display.FormatNumber(avg)))
		}
	})
	fmt.Println()

	if len(tokensByModel) > 0 {
		display.Box("Tokens (output)", func() {
			for model, tokens := range tokensByModel {
				fmt.Printf("  %-14s  %s\n",
					display.ModelShort(model),
					display.BoldWhite(display.FormatTokens(tokens)))
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

	// Load sessions from indexes for this period
	sessions, err := store.ListSessionsAfter(startDate)
	if err == nil && len(sessions) > 0 {
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
