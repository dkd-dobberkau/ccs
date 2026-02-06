package store

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dkd/ccs/internal/claude"
)

// LoadStatsCache reads and parses ~/.claude/stats-cache.json
func LoadStatsCache() (*StatsCache, error) {
	path := claude.StatsCache()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var stats StatsCache
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return &stats, nil
}

// SaveStatsCache writes the stats cache to ~/.claude/stats-cache.json
func SaveStatsCache(stats *StatsCache) error {
	path := claude.StatsCache()
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling stats: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
