package store

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/dkd/ccs/internal/claude"
)

// LoadHistory reads the global history.jsonl and returns entries
func LoadHistory(limit int) ([]HistoryEntry, error) {
	path := claude.HistoryFile()
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []HistoryEntry
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 256*1024)
	scanner.Buffer(buf, 2*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry HistoryEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		// Only include entries with display text (user prompts)
		if entry.Display != "" {
			entries = append(entries, entry)
		}
	}

	// Return last N entries
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}

	return entries, scanner.Err()
}
