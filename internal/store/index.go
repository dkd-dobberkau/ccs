package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dkd/ccs/internal/claude"
)

// LoadAllProjects scans all project directories and returns aggregated project info
func LoadAllProjects() ([]Project, error) {
	projectsDir := claude.ProjectsDir()
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	var projects []Project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		projDir := filepath.Join(projectsDir, dirName)

		p := Project{
			DirName: dirName,
		}

		// Try to load sessions-index.json
		indexPath := filepath.Join(projDir, "sessions-index.json")
		if idx, err := loadSessionIndex(indexPath); err == nil {
			p.HasIndex = true
			p.Path = idx.OriginalPath
			p.SessionCount = len(idx.Entries)
			for _, e := range idx.Entries {
				p.MessageCount += e.MessageCount
				if t, err := time.Parse(time.RFC3339, e.Modified); err == nil {
					if t.After(p.LastActive) {
						p.LastActive = t
					}
				}
			}
		} else {
			// Fallback: count JSONL files and use mtime
			// Don't attempt path conversion - dash encoding is ambiguous
			p.Path = dirName
			jsonlFiles, _ := filepath.Glob(filepath.Join(projDir, "*.jsonl"))
			p.SessionCount = len(jsonlFiles)
			for _, f := range jsonlFiles {
				if info, err := os.Stat(f); err == nil {
					if info.ModTime().After(p.LastActive) {
						p.LastActive = info.ModTime()
					}
				}
			}
		}

		// Skip empty projects
		if p.SessionCount == 0 {
			continue
		}

		projects = append(projects, p)
	}

	return projects, nil
}

// ListAllSessions returns all sessions across all projects, sorted by created desc
func ListAllSessions(projectFilter string) ([]SessionEntry, error) {
	projectsDir := claude.ProjectsDir()
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	var allSessions []SessionEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()

		// Apply project filter: match against dir name and originalPath from index
		if projectFilter != "" {
			filter := strings.ToLower(projectFilter)
			if !strings.Contains(strings.ToLower(dirName), filter) {
				// Also check sessions-index for originalPath match
				idxPath := filepath.Join(projectsDir, dirName, "sessions-index.json")
				if idx, err := loadSessionIndex(idxPath); err != nil || !strings.Contains(strings.ToLower(idx.OriginalPath), filter) {
					continue
				}
			}
		}

		indexPath := filepath.Join(projectsDir, dirName, "sessions-index.json")
		idx, err := loadSessionIndex(indexPath)
		if err != nil {
			continue
		}

		allSessions = append(allSessions, idx.Entries...)
	}

	// Sort by created descending
	sort.Slice(allSessions, func(i, j int) bool {
		return allSessions[i].Created > allSessions[j].Created
	})

	return allSessions, nil
}

// ListSessionsAfter returns all sessions created after the given time
func ListSessionsAfter(after time.Time) ([]SessionEntry, error) {
	sessions, err := ListAllSessions("")
	if err != nil {
		return nil, err
	}

	afterStr := after.Format(time.RFC3339)
	var filtered []SessionEntry
	for _, s := range sessions {
		if s.Created >= afterStr {
			filtered = append(filtered, s)
		}
	}

	return filtered, nil
}

// FindSession finds a session by ID prefix match, returns path and entry
func FindSession(idPrefix string) (string, *SessionEntry, error) {
	projectsDir := claude.ProjectsDir()
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return "", nil, err
	}

	idPrefix = strings.ToLower(idPrefix)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projDir := filepath.Join(projectsDir, entry.Name())

		// Check sessions-index.json first for metadata
		indexPath := filepath.Join(projDir, "sessions-index.json")
		if idx, err := loadSessionIndex(indexPath); err == nil {
			for i, e := range idx.Entries {
				if strings.HasPrefix(strings.ToLower(e.SessionID), idPrefix) {
					return e.FullPath, &idx.Entries[i], nil
				}
			}
		}

		// Fallback: match JSONL filenames
		jsonlFiles, _ := filepath.Glob(filepath.Join(projDir, "*.jsonl"))
		for _, f := range jsonlFiles {
			base := strings.TrimSuffix(filepath.Base(f), ".jsonl")
			if strings.HasPrefix(strings.ToLower(base), idPrefix) {
				return f, nil, nil
			}
		}
	}

	return "", nil, os.ErrNotExist
}

func loadSessionIndex(path string) (*SessionIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var idx SessionIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

