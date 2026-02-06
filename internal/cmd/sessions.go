package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dkd/ccs/internal/display"
	"github.com/dkd/ccs/internal/store"
)

func Sessions(args []string) error {
	var project string
	limit := 20

	for i := 0; i < len(args); i++ {
		switch {
		case strings.HasPrefix(args[i], "--project="):
			project = strings.TrimPrefix(args[i], "--project=")
		case args[i] == "-n" && i+1 < len(args):
			n, err := strconv.Atoi(args[i+1])
			if err == nil {
				limit = n
			}
			i++
		}
	}

	sessions, err := store.ListAllSessions(project)
	if err != nil {
		return fmt.Errorf("loading sessions: %w", err)
	}

	if display.IsJSON() {
		return sessionsJSON(sessions, limit)
	}
	if display.IsMD() {
		return sessionsMD(sessions, project, limit)
	}

	title := "Recent Sessions"
	if project != "" {
		title = fmt.Sprintf("Sessions for %s", project)
	}
	fmt.Println(display.BoldCyan(title))
	fmt.Printf("Total: %d sessions\n\n", len(sessions))

	if len(sessions) == 0 {
		fmt.Println(display.Dim("No sessions found"))
		return nil
	}

	if limit > len(sessions) {
		limit = len(sessions)
	}

	for i := 0; i < limit; i++ {
		s := sessions[i]
		prompt := display.Truncate(s.FirstPrompt, 55)
		if prompt == "" {
			prompt = display.Dim("(no prompt)")
		}

		created := ""
		if t, err := time.Parse(time.RFC3339, s.Created); err == nil {
			created = display.RelativeTime(t)
		}

		sidechain := ""
		if s.IsSidechain {
			sidechain = display.Yellow(" [sidechain]")
		}

		branch := ""
		if s.GitBranch != "" {
			branch = display.Dim(" (" + s.GitBranch + ")")
		}

		fmt.Printf("  %s  %3d msgs  %-10s  %s%s%s\n",
			display.Dim(s.SessionID[:8]),
			s.MessageCount,
			created,
			prompt,
			sidechain,
			branch)
	}

	if len(sessions) > limit {
		fmt.Printf("\n  %s\n", display.Dim(fmt.Sprintf("Showing %d of %d. Use -n to see more.", limit, len(sessions))))
	}
	fmt.Println()

	return nil
}

func sessionsJSON(sessions []store.SessionEntry, limit int) error {
	if limit > len(sessions) {
		limit = len(sessions)
	}
	type jsonSession struct {
		SessionID  string `json:"sessionId"`
		Messages   int    `json:"messages"`
		Created    string `json:"created"`
		FirstPrompt string `json:"firstPrompt"`
		Branch     string `json:"branch,omitempty"`
		Sidechain  bool   `json:"sidechain,omitempty"`
	}
	out := make([]jsonSession, 0, limit)
	for i := 0; i < limit; i++ {
		s := sessions[i]
		out = append(out, jsonSession{
			SessionID:  s.SessionID,
			Messages:   s.MessageCount,
			Created:    s.Created,
			FirstPrompt: s.FirstPrompt,
			Branch:     s.GitBranch,
			Sidechain:  s.IsSidechain,
		})
	}
	return OutputJSON(out)
}

func sessionsMD(sessions []store.SessionEntry, project string, limit int) error {
	title := "Recent Sessions"
	if project != "" {
		title = fmt.Sprintf("Sessions for %s", project)
	}
	MDHeader(2, title)
	fmt.Printf("Total: %d sessions\n\n", len(sessions))

	if len(sessions) == 0 {
		fmt.Println("No sessions found.")
		fmt.Println()
		return nil
	}

	if limit > len(sessions) {
		limit = len(sessions)
	}

	headers := []string{"ID", "Messages", "Created", "Prompt", "Branch"}
	var rows [][]string
	for i := 0; i < limit; i++ {
		s := sessions[i]
		prompt := display.Truncate(s.FirstPrompt, 55)
		if prompt == "" {
			prompt = "(no prompt)"
		}
		created := ""
		if t, err := time.Parse(time.RFC3339, s.Created); err == nil {
			created = display.RelativeTime(t)
		}
		sidechain := ""
		if s.IsSidechain {
			sidechain = " [sidechain]"
		}
		rows = append(rows, []string{
			s.SessionID[:8],
			fmt.Sprintf("%d", s.MessageCount),
			created,
			prompt + sidechain,
			s.GitBranch,
		})
	}
	MDTable(headers, rows)

	if len(sessions) > limit {
		fmt.Printf("Showing %d of %d. Use -n to see more.\n\n", limit, len(sessions))
	}
	return nil
}
