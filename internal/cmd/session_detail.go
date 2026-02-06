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

	if display.IsJSON() {
		return sessionDetailJSON(detail, entry)
	}
	if display.IsMD() {
		return sessionDetailMD(detail, entry)
	}

	fmt.Println(display.BoldCyan("Session Detail"))
	fmt.Println()

	display.Box("Info", func() {
		fmt.Printf("  ID          %s\n", display.Bold(detail.ID))
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
			fmt.Printf("  Duration    %s\n", display.Bold(display.FormatDurationFromTime(duration)))
		}
		if detail.Model != "" {
			fmt.Printf("  Model       %s\n", display.Bold(display.ModelShort(detail.Model)))
		}
		if detail.IsSidechain {
			fmt.Printf("  Type        %s\n", display.Yellow("sidechain"))
		}
	})
	fmt.Println()

	display.Box("Messages", func() {
		fmt.Printf("  Total       %s\n", display.Bold(display.FormatNumber(detail.TotalMessages)))
		fmt.Printf("  User        %s\n", display.FormatNumber(detail.UserMessages))
		fmt.Printf("  Assistant   %s\n", display.FormatNumber(detail.AsstMessages))
	})
	fmt.Println()

	display.Box("Tokens", func() {
		fmt.Printf("  Input       %s\n", display.Bold(display.FormatTokens(detail.TotalTokensIn)))
		fmt.Printf("  Output      %s\n", display.Bold(display.FormatTokens(detail.TotalTokensOut)))
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

func sessionDetailJSON(detail *store.SessionDetail, entry *store.SessionEntry) error {
	tools := make(map[string]int)
	for name, stats := range detail.Tools {
		tools[name] = stats.Count
	}

	type jsonMessage struct {
		Time    string `json:"time,omitempty"`
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	var conversation []jsonMessage
	for _, msg := range detail.Messages {
		if msg.Content == "" {
			continue
		}
		ts := ""
		if !msg.Timestamp.IsZero() {
			ts = msg.Timestamp.Format(time.RFC3339)
		}
		conversation = append(conversation, jsonMessage{
			Time:    ts,
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	data := map[string]any{
		"id":       detail.ID,
		"model":    display.ModelShort(detail.Model),
		"messages": map[string]int{
			"total":     detail.TotalMessages,
			"user":      detail.UserMessages,
			"assistant": detail.AsstMessages,
		},
		"tokens": map[string]int{
			"input":  detail.TotalTokensIn,
			"output": detail.TotalTokensOut,
		},
		"tools":        tools,
		"conversation": conversation,
	}

	if entry != nil && entry.ProjectPath != "" {
		data["project"] = entry.ProjectPath
	}
	if detail.GitBranch != "" {
		data["branch"] = detail.GitBranch
	}
	if !detail.StartedAt.IsZero() {
		data["started"] = detail.StartedAt.Format(time.RFC3339)
	}
	if !detail.EndedAt.IsZero() {
		duration := detail.EndedAt.Sub(detail.StartedAt)
		data["duration"] = display.FormatDurationFromTime(duration)
	}

	return OutputJSON(data)
}

func sessionDetailMD(detail *store.SessionDetail, entry *store.SessionEntry) error {
	MDHeader(2, "Session Detail")

	MDHeader(3, "Info")
	fmt.Printf("- **ID:** %s\n", detail.ID)
	if entry != nil && entry.ProjectPath != "" {
		fmt.Printf("- **Project:** %s\n", entry.ProjectPath)
	}
	if detail.GitBranch != "" {
		fmt.Printf("- **Branch:** %s\n", detail.GitBranch)
	}
	if detail.Version != "" {
		fmt.Printf("- **CLI:** v%s\n", detail.Version)
	}
	if !detail.StartedAt.IsZero() {
		fmt.Printf("- **Started:** %s\n", detail.StartedAt.Format(time.RFC3339))
	}
	if !detail.EndedAt.IsZero() {
		duration := detail.EndedAt.Sub(detail.StartedAt)
		fmt.Printf("- **Duration:** %s\n", display.FormatDurationFromTime(duration))
	}
	if detail.Model != "" {
		fmt.Printf("- **Model:** %s\n", display.ModelShort(detail.Model))
	}
	fmt.Println()

	MDHeader(3, "Messages")
	fmt.Printf("- **Total:** %s\n", display.FormatNumber(detail.TotalMessages))
	fmt.Printf("- **User:** %s\n", display.FormatNumber(detail.UserMessages))
	fmt.Printf("- **Assistant:** %s\n\n", display.FormatNumber(detail.AsstMessages))

	MDHeader(3, "Tokens")
	fmt.Printf("- **Input:** %s\n", display.FormatTokens(detail.TotalTokensIn))
	fmt.Printf("- **Output:** %s\n\n", display.FormatTokens(detail.TotalTokensOut))

	if len(detail.Tools) > 0 {
		MDHeader(3, "Tool Usage")
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
		tHeaders := []string{"Tool", "Count"}
		var tRows [][]string
		for _, t := range tools {
			tRows = append(tRows, []string{t.name, fmt.Sprintf("%d", t.count)})
		}
		MDTable(tHeaders, tRows)
	}

	MDHeader(3, "Conversation")
	shown := 0
	for _, msg := range detail.Messages {
		if msg.Role != "user" || msg.Content == "" {
			continue
		}
		if shown >= 10 {
			fmt.Println("... and more messages")
			fmt.Println()
			break
		}
		ts := ""
		if !msg.Timestamp.IsZero() {
			ts = msg.Timestamp.Format("15:04") + " "
		}
		prompt := display.Truncate(msg.Content, 70)
		fmt.Printf("- %s**>** %s\n", ts, prompt)
		shown++
	}
	if shown == 0 {
		fmt.Println("(no user messages)")
	}
	fmt.Println()
	return nil
}
