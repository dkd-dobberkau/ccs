package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/dkd/ccs/internal/display"
	"github.com/dkd/ccs/internal/store"
)

func SessionDetail(idPrefix string) error {
	// Find session file by prefix match
	path, entry, err := store.FindSession(idPrefix)
	if err != nil {
		return fmt.Errorf("finding session: %w", err)
	}

	// Parse the full JSONL file
	detail, err := store.ParseSessionJSONL(path)
	if err != nil {
		return fmt.Errorf("parsing session: %w", err)
	}

	fmt.Println(display.BoldCyan("Session Detail"))
	fmt.Println()

	display.Box("Info", func() {
		fmt.Printf("  ID          %s\n", display.BoldWhite(detail.ID))
		if entry != nil && entry.ProjectPath != "" {
			fmt.Printf("  Project     %s\n", entry.ProjectPath)
		}
		if detail.GitBranch != "" {
			fmt.Printf("  Branch      %s\n", detail.GitBranch)
		}
		if detail.Version != "" {
			fmt.Printf("  CLI         %s\n", display.Dim("v"+detail.Version))
		}
		if !detail.StartedAt.IsZero() {
			fmt.Printf("  Started     %s\n", detail.StartedAt.Format(time.RFC3339))
		}
		if !detail.EndedAt.IsZero() {
			duration := detail.EndedAt.Sub(detail.StartedAt)
			fmt.Printf("  Duration    %s\n", display.BoldWhite(display.FormatDurationFromTime(duration)))
		}
		if detail.Model != "" {
			fmt.Printf("  Model       %s\n", display.BoldWhite(display.ModelShort(detail.Model)))
		}
		if detail.IsSidechain {
			fmt.Printf("  Type        %s\n", display.Yellow("sidechain"))
		}
	})
	fmt.Println()

	display.Box("Messages", func() {
		fmt.Printf("  Total       %s\n", display.BoldWhite(display.FormatNumber(detail.TotalMessages)))
		fmt.Printf("  User        %s\n", display.FormatNumber(detail.UserMessages))
		fmt.Printf("  Assistant   %s\n", display.FormatNumber(detail.AsstMessages))
	})
	fmt.Println()

	display.Box("Tokens", func() {
		fmt.Printf("  Input       %s\n", display.BoldWhite(display.FormatTokens(detail.TotalTokensIn)))
		fmt.Printf("  Output      %s\n", display.BoldWhite(display.FormatTokens(detail.TotalTokensOut)))
	})
	fmt.Println()

	// Tool usage
	if len(detail.Tools) > 0 {
		type toolEntry struct {
			name  string
			count int
		}
		var tools []toolEntry
		for name, stats := range detail.Tools {
			tools = append(tools, toolEntry{name, stats.Count})
		}
		sort.Slice(tools, func(i, j int) bool {
			return tools[i].count > tools[j].count
		})

		maxCount := tools[0].count
		display.Box("Tool Usage", func() {
			for _, t := range tools {
				bar := display.Bar(t.count, maxCount, 15)
				fmt.Printf("  %s %3d  %s\n", bar, t.count, t.name)
			}
		})
		fmt.Println()
	}

	// First few user messages
	display.Box("Conversation", func() {
		shown := 0
		for _, msg := range detail.Messages {
			if msg.Role != "user" || msg.Content == "" {
				continue
			}
			if shown >= 10 {
				fmt.Printf("  %s\n", display.Dim(fmt.Sprintf("... and more messages")))
				break
			}
			ts := ""
			if !msg.Timestamp.IsZero() {
				ts = msg.Timestamp.Format("15:04") + " "
			}
			prompt := display.Truncate(msg.Content, 70)
			fmt.Printf("  %s%s %s\n",
				display.Dim(ts),
				display.Green("▸"),
				prompt)
			shown++
		}
		if shown == 0 {
			fmt.Printf("  %s\n", display.Dim("(no user messages)"))
		}
	})
	fmt.Println()

	return nil
}
