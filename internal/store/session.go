package store

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ParseSessionJSONL parses a full session JSONL file into a SessionDetail
func ParseSessionJSONL(path string) (*SessionDetail, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	detail := &SessionDetail{
		ID:    strings.TrimSuffix(filepath.Base(path), ".jsonl"),
		Tools: make(map[string]*ToolStats),
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var firstTs, lastTs time.Time
	msgSeq := 0

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry RawEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		// Capture session metadata from first entry
		if entry.SessionID != "" && detail.ID == "" {
			detail.ID = entry.SessionID
		}
		if entry.Version != "" {
			detail.Version = entry.Version
		}
		if entry.GitBranch != "" {
			detail.GitBranch = entry.GitBranch
		}
		if entry.IsSidechain {
			detail.IsSidechain = true
		}

		// Parse timestamp
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
			detail.TotalMessages++
			detail.UserMessages++

			var msgContent MessageContent
			if entry.Message != nil {
				json.Unmarshal(entry.Message, &msgContent)
			}

			msgTs, _ := time.Parse(time.RFC3339, entry.Timestamp)

			// Extract text from user message
			var text string
			if len(msgContent.Content) > 0 {
				var contentStr string
				if err := json.Unmarshal(msgContent.Content, &contentStr); err == nil {
					text = contentStr
				} else {
					// Could be content blocks
					var blocks []ContentBlock
					if err := json.Unmarshal(msgContent.Content, &blocks); err == nil {
						var parts []string
						for _, b := range blocks {
							if b.Type == "text" && b.Text != "" {
								parts = append(parts, b.Text)
							}
						}
						text = strings.Join(parts, "\n")
					}
				}
			}

			detail.Messages = append(detail.Messages, Message{
				Seq:       msgSeq,
				Timestamp: msgTs,
				Role:      "user",
				Content:   text,
			})
			msgSeq++

		case "assistant":
			detail.TotalMessages++
			detail.AsstMessages++

			var msgContent MessageContent
			if entry.Message != nil {
				json.Unmarshal(entry.Message, &msgContent)
			}

			msgTs, _ := time.Parse(time.RFC3339, entry.Timestamp)

			// Extract text and tools from assistant message
			var text string
			if len(msgContent.Content) > 0 {
				var blocks []ContentBlock
				if err := json.Unmarshal(msgContent.Content, &blocks); err == nil {
					var parts []string
					for _, block := range blocks {
						switch block.Type {
						case "text":
							parts = append(parts, block.Text)
						case "tool_use":
							if block.Name != "" {
								if detail.Tools[block.Name] == nil {
									detail.Tools[block.Name] = &ToolStats{}
								}
								detail.Tools[block.Name].Count++
							}
						}
					}
					text = strings.Join(parts, "\n")
				}
			}

			// Track tokens and model
			if msgContent.Usage != nil {
				detail.TotalTokensIn += msgContent.Usage.InputTokens
				detail.TotalTokensOut += msgContent.Usage.OutputTokens
			}
			if msgContent.Model != "" {
				detail.Model = msgContent.Model
			}

			detail.Messages = append(detail.Messages, Message{
				Seq:       msgSeq,
				Timestamp: msgTs,
				Role:      "assistant",
				Content:   text,
			})
			msgSeq++
		}
	}

	if !firstTs.IsZero() {
		detail.StartedAt = firstTs
	}
	if !lastTs.IsZero() {
		detail.EndedAt = lastTs
	}

	return detail, scanner.Err()
}
