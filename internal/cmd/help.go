package cmd

import "fmt"

func Help(version string) {
	fmt.Printf(`ccs %s - Claude Code Summary

Usage: ccs [command] [flags]

Commands:
  all              Full report (summary + projects + sessions + tokens)
  summary          Dashboard overview (default)
  today            Today's activity
  week             This week's activity
  month            This month's activity
  projects         Project ranking by activity
  sessions         List recent sessions
  session <id>     Session detail view
  tokens           Token usage breakdown
  version          Show version
  help             Show this help

Global flags:
  --json           Output as JSON
  --md             Output as Markdown

Flags (sessions):
  --project=X      Filter by project name
  -n N             Limit number of results (default: 20)

Data source: ~/.claude/
`, version)
}
