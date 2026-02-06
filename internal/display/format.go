package display

import (
	"fmt"
	"strings"
	"time"
)

// FormatNumber adds thousand separators: 1234567 → "1,234,567"
func FormatNumber(n int) string {
	if n < 0 {
		return "-" + FormatNumber(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, ",")
}

// FormatDuration formats milliseconds into human-readable duration
func FormatDuration(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

// FormatDurationFromTime formats a time.Duration into human-readable string
func FormatDurationFromTime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

// RelativeTime formats a time relative to now: "2h ago", "3d ago"
func RelativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh ago", h)
	case d < 30*24*time.Hour:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("2006-01-02")
	}
}

// ModelShort returns a short display name for a model ID
func ModelShort(model string) string {
	switch {
	case strings.Contains(model, "opus-4-6"):
		return "Opus 4.6"
	case strings.Contains(model, "opus-4-5"):
		return "Opus 4.5"
	case strings.Contains(model, "sonnet-4-5"):
		return "Sonnet 4.5"
	case strings.Contains(model, "haiku-4-5"):
		return "Haiku 4.5"
	case strings.Contains(model, "opus"):
		return "Opus"
	case strings.Contains(model, "sonnet"):
		return "Sonnet"
	case strings.Contains(model, "haiku"):
		return "Haiku"
	default:
		return model
	}
}

// Truncate truncates a string to maxLen, appending "..." if truncated
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// FormatTokens formats token count with K/M suffix
func FormatTokens(n int) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
