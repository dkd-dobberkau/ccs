package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dkd/ccs/internal/claude"
)

// SessionStats holds lightweight stats extracted from a session JSONL file.
type SessionStats struct {
	SessionID    string
	StartedAt    time.Time
	EndedAt      time.Time
	UserMessages int
	AsstMessages int
	ToolCalls    int
	TokensIn     map[string]int // keyed by model
	TokensOut    map[string]int
	CacheRead    map[string]int
	CacheCreate  map[string]int
	Model        string
}

// ScanSessionStats does a lightweight parse of a session JSONL file,
// extracting only stats (tokens, model, tool counts, timestamps) without
// building full message content or the Messages slice.
func ScanSessionStats(path string) (*SessionStats, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ss := &SessionStats{
		SessionID:   strings.TrimSuffix(filepath.Base(path), ".jsonl"),
		TokensIn:    make(map[string]int),
		TokensOut:   make(map[string]int),
		CacheRead:   make(map[string]int),
		CacheCreate: make(map[string]int),
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 256*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var firstTs, lastTs time.Time

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry RawEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		if entry.Timestamp != "" {
			if ts, err := time.Parse(time.RFC3339, entry.Timestamp); err == nil {
				if firstTs.IsZero() {
					firstTs = ts
				}
				lastTs = ts
			}
		}

		switch entry.Type {
		case "user":
			ss.UserMessages++
		case "assistant":
			ss.AsstMessages++
			if entry.Message == nil {
				continue
			}

			var msgContent MessageContent
			if err := json.Unmarshal(entry.Message, &msgContent); err != nil {
				continue
			}

			model := msgContent.Model
			if model != "" {
				ss.Model = model
			}

			if msgContent.Usage != nil {
				m := ss.Model
				if m == "" {
					m = "unknown"
				}
				ss.TokensIn[m] += msgContent.Usage.InputTokens
				ss.TokensOut[m] += msgContent.Usage.OutputTokens
				ss.CacheRead[m] += msgContent.Usage.CacheReadInputTokens
				ss.CacheCreate[m] += msgContent.Usage.CacheCreationInputTokens
			}

			// Count tool_use blocks without extracting text
			if len(msgContent.Content) > 0 {
				var blocks []ContentBlock
				if err := json.Unmarshal(msgContent.Content, &blocks); err == nil {
					for _, b := range blocks {
						if b.Type == "tool_use" {
							ss.ToolCalls++
						}
					}
				}
			}
		}
	}

	ss.StartedAt = firstTs
	ss.EndedAt = lastTs

	return ss, scanner.Err()
}

// ComputeStats scans all session JSONL files and builds a fresh StatsCache.
// The progress callback is called after each session is scanned.
func ComputeStats(progress func(done, total int)) (*StatsCache, error) {
	projectsDir := claude.ProjectsDir()
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, fmt.Errorf("reading projects dir: %w", err)
	}

	// Collect all JSONL file paths
	var allFiles []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projDir := filepath.Join(projectsDir, entry.Name())
		files, _ := filepath.Glob(filepath.Join(projDir, "*.jsonl"))
		allFiles = append(allFiles, files...)
	}

	total := len(allFiles)

	// Aggregation maps
	dailyAct := make(map[string]*DailyActivity)    // date -> activity
	dailyTokens := make(map[string]map[string]int)  // date -> model -> output tokens
	modelUsage := make(map[string]*ModelUsage)       // model -> usage
	hourCounts := make(map[string]int)               // hour string -> count

	var totalSessions, totalMessages int
	var firstSession string
	var longest LongestSession

	for i, path := range allFiles {
		if progress != nil {
			progress(i+1, total)
		}

		ss, err := ScanSessionStats(path)
		if err != nil {
			continue
		}

		msgCount := ss.UserMessages + ss.AsstMessages
		if msgCount == 0 {
			continue
		}

		totalSessions++
		totalMessages += msgCount

		// Date from first timestamp
		date := ""
		if !ss.StartedAt.IsZero() {
			date = ss.StartedAt.Format("2006-01-02")
		}

		// First session date
		if date != "" && (firstSession == "" || date < firstSession) {
			firstSession = date
		}

		// Hour counts
		if !ss.StartedAt.IsZero() {
			h := fmt.Sprintf("%d", ss.StartedAt.Hour())
			hourCounts[h]++
		}

		// Daily activity
		if date != "" {
			da, ok := dailyAct[date]
			if !ok {
				da = &DailyActivity{Date: date}
				dailyAct[date] = da
			}
			da.SessionCount++
			da.MessageCount += msgCount
			da.ToolCallCount += ss.ToolCalls
		}

		// Token aggregation per model
		for model, out := range ss.TokensOut {
			// Daily tokens
			if date != "" {
				if dailyTokens[date] == nil {
					dailyTokens[date] = make(map[string]int)
				}
				dailyTokens[date][model] += out
			}

			// Model totals
			mu, ok := modelUsage[model]
			if !ok {
				mu = &ModelUsage{}
				modelUsage[model] = mu
			}
			mu.OutputTokens += out
			mu.InputTokens += ss.TokensIn[model]
			mu.CacheReadInputTokens += ss.CacheRead[model]
			mu.CacheCreationInputTokens += ss.CacheCreate[model]
		}

		// Longest session
		if !ss.StartedAt.IsZero() && !ss.EndedAt.IsZero() {
			dur := ss.EndedAt.Sub(ss.StartedAt).Milliseconds()
			if dur > longest.Duration {
				longest = LongestSession{
					SessionID:    ss.SessionID,
					Duration:     dur,
					MessageCount: msgCount,
					Timestamp:    ss.StartedAt.Format(time.RFC3339),
				}
			}
		}
	}

	// Build sorted DailyActivity slice
	var dailyActivitySlice []DailyActivity
	for _, da := range dailyAct {
		dailyActivitySlice = append(dailyActivitySlice, *da)
	}
	sort.Slice(dailyActivitySlice, func(i, j int) bool {
		return dailyActivitySlice[i].Date < dailyActivitySlice[j].Date
	})

	// Build sorted DailyModelTokens slice
	var dailyModelTokensSlice []DailyModelTokens
	for date, byModel := range dailyTokens {
		dailyModelTokensSlice = append(dailyModelTokensSlice, DailyModelTokens{
			Date:          date,
			TokensByModel: byModel,
		})
	}
	sort.Slice(dailyModelTokensSlice, func(i, j int) bool {
		return dailyModelTokensSlice[i].Date < dailyModelTokensSlice[j].Date
	})

	// Build ModelUsage map with string keys
	muMap := make(map[string]ModelUsage)
	for model, mu := range modelUsage {
		muMap[model] = *mu
	}

	// Handle firstSession as RFC3339 if it's just a date
	if firstSession != "" && !strings.Contains(firstSession, "T") {
		firstSession = firstSession + "T00:00:00Z"
	}

	stats := &StatsCache{
		Version:          1,
		LastComputedDate: time.Now().Format("2006-01-02"),
		DailyActivity:    dailyActivitySlice,
		DailyModelTokens: dailyModelTokensSlice,
		ModelUsage:       muMap,
		TotalSessions:    totalSessions,
		TotalMessages:    totalMessages,
		LongestSession:   longest,
		FirstSessionDate: firstSession,
		HourCounts:       hourCounts,
	}

	return stats, nil
}
