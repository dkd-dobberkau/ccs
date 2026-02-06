package store

import (
	"encoding/json"
	"time"
)

// StatsCache represents the top-level stats-cache.json structure
type StatsCache struct {
	Version          int              `json:"version"`
	LastComputedDate string           `json:"lastComputedDate"`
	DailyActivity    []DailyActivity  `json:"dailyActivity"`
	DailyModelTokens []DailyModelTokens `json:"dailyModelTokens"`
	ModelUsage       map[string]ModelUsage `json:"modelUsage"`
	TotalSessions    int              `json:"totalSessions"`
	TotalMessages    int              `json:"totalMessages"`
	LongestSession   LongestSession   `json:"longestSession"`
	FirstSessionDate string           `json:"firstSessionDate"`
	HourCounts       map[string]int   `json:"hourCounts"`
}

type DailyActivity struct {
	Date         string `json:"date"`
	MessageCount int    `json:"messageCount"`
	SessionCount int    `json:"sessionCount"`
	ToolCallCount int   `json:"toolCallCount"`
}

type DailyModelTokens struct {
	Date          string         `json:"date"`
	TokensByModel map[string]int `json:"tokensByModel"`
}

type ModelUsage struct {
	InputTokens              int `json:"inputTokens"`
	OutputTokens             int `json:"outputTokens"`
	CacheReadInputTokens     int `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int `json:"cacheCreationInputTokens"`
}

type LongestSession struct {
	SessionID    string `json:"sessionId"`
	Duration     int64  `json:"duration"`
	MessageCount int    `json:"messageCount"`
	Timestamp    string `json:"timestamp"`
}

// SessionIndex represents a project's sessions-index.json
type SessionIndex struct {
	Version      int            `json:"version"`
	Entries      []SessionEntry `json:"entries"`
	OriginalPath string         `json:"originalPath"`
}

type SessionEntry struct {
	SessionID   string `json:"sessionId"`
	FullPath    string `json:"fullPath"`
	FileMtime   int64  `json:"fileMtime"`
	FirstPrompt string `json:"firstPrompt"`
	MessageCount int   `json:"messageCount"`
	Created     string `json:"created"`
	Modified    string `json:"modified"`
	GitBranch   string `json:"gitBranch"`
	ProjectPath string `json:"projectPath"`
	IsSidechain bool   `json:"isSidechain"`
}

// Project aggregates session data for a project directory
type Project struct {
	DirName      string
	Path         string // originalPath from index, or derived
	SessionCount int
	MessageCount int
	LastActive   time.Time
	HasIndex     bool
}

// HistoryEntry represents a line in history.jsonl
type HistoryEntry struct {
	Display  string `json:"display"`
	Timestamp int64 `json:"timestamp"`
	Project  string `json:"project"`
}

// SessionDetail represents a fully parsed session JSONL file
type SessionDetail struct {
	ID            string
	ProjectName   string
	StartedAt     time.Time
	EndedAt       time.Time
	TotalMessages int
	UserMessages  int
	AsstMessages  int
	TotalTokensIn int
	TotalTokensOut int
	Model         string
	Tools         map[string]*ToolStats
	Messages      []Message
	GitBranch     string
	Version       string
	IsSidechain   bool
}

type ToolStats struct {
	Count int
}

type Message struct {
	Seq       int
	Timestamp time.Time
	Role      string
	Content   string
}

// RawEntry represents a single line in a session JSONL file
type RawEntry struct {
	Type        string          `json:"type"`
	Timestamp   string          `json:"timestamp,omitempty"`
	Message     json.RawMessage `json:"message,omitempty"`
	SessionID   string          `json:"sessionId,omitempty"`
	Version     string          `json:"version,omitempty"`
	GitBranch   string          `json:"gitBranch,omitempty"`
	IsSidechain bool            `json:"isSidechain"`
	CWD         string          `json:"cwd,omitempty"`
}

type MessageContent struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Usage   *Usage          `json:"usage,omitempty"`
	Model   string          `json:"model,omitempty"`
}

type ContentBlock struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	Name  string `json:"name,omitempty"`
	Input any    `json:"input,omitempty"`
}

type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}
