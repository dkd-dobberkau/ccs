package claude

import (
	"os"
	"path/filepath"
)

// Dir returns the path to ~/.claude/
func Dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude")
}

// StatsCache returns the path to stats-cache.json
func StatsCache() string {
	return filepath.Join(Dir(), "stats-cache.json")
}

// ProjectsDir returns the path to the projects directory
func ProjectsDir() string {
	return filepath.Join(Dir(), "projects")
}

// HistoryFile returns the path to history.jsonl
func HistoryFile() string {
	return filepath.Join(Dir(), "history.jsonl")
}
