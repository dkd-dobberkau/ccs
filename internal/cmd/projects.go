package cmd

import (
	"fmt"
	"sort"

	"github.com/dkd/ccs/internal/display"
	"github.com/dkd/ccs/internal/store"
)

func Projects() error {
	projects, err := store.LoadAllProjects()
	if err != nil {
		return fmt.Errorf("loading projects: %w", err)
	}

	// Sort by message count descending
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].MessageCount > projects[j].MessageCount
	})

	fmt.Println(display.BoldCyan("Projects"))
	fmt.Printf("Found %d projects with activity\n\n", len(projects))

	if len(projects) == 0 {
		fmt.Println(display.Dim("No projects found"))
		return nil
	}

	// Find max for bar chart
	maxMsgs := projects[0].MessageCount

	for _, p := range projects {
		name := p.Path
		if name == "" {
			name = p.DirName
		}
		bar := display.Bar(p.MessageCount, maxMsgs, 15)
		lastActive := display.RelativeTime(p.LastActive)
		fmt.Printf("  %s %-6s  %3d sessions  %s  %s\n",
			bar,
			display.FormatNumber(p.MessageCount)+" msgs",
			p.SessionCount,
			display.Dim(fmt.Sprintf("%-10s", lastActive)),
			display.BoldWhite(name))
	}
	fmt.Println()

	return nil
}
