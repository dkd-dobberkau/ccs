package cmd

import (
	"fmt"
	"sort"
	"time"

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

	if display.IsJSON() {
		return projectsJSON(projects)
	}
	if display.IsMD() {
		return projectsMD(projects)
	}

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
			display.Bold(name))
	}
	fmt.Println()

	return nil
}

func projectsJSON(projects []store.Project) error {
	type jsonProject struct {
		Path       string `json:"path"`
		Sessions   int    `json:"sessions"`
		Messages   int    `json:"messages"`
		LastActive string `json:"lastActive"`
	}
	out := make([]jsonProject, 0, len(projects))
	for _, p := range projects {
		name := p.Path
		if name == "" {
			name = p.DirName
		}
		la := ""
		if !p.LastActive.IsZero() {
			la = p.LastActive.Format(time.RFC3339)
		}
		out = append(out, jsonProject{
			Path:       name,
			Sessions:   p.SessionCount,
			Messages:   p.MessageCount,
			LastActive: la,
		})
	}
	return OutputJSON(out)
}

func projectsMD(projects []store.Project) error {
	MDHeader(2, "Projects")
	fmt.Printf("Found %d projects with activity\n\n", len(projects))

	if len(projects) == 0 {
		fmt.Println("No projects found.")
		fmt.Println()
		return nil
	}

	headers := []string{"Project", "Sessions", "Messages", "Last Active"}
	var rows [][]string
	for _, p := range projects {
		name := p.Path
		if name == "" {
			name = p.DirName
		}
		lastActive := display.RelativeTime(p.LastActive)
		rows = append(rows, []string{
			name,
			fmt.Sprintf("%d", p.SessionCount),
			display.FormatNumber(p.MessageCount),
			lastActive,
		})
	}
	MDTable(headers, rows)
	return nil
}
