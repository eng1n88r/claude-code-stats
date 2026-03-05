package extract

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type historyEntry struct {
	Display   string
	Timestamp float64
	Project   string
	SessionID string
}

// LoadHistory loads history.jsonl from primary + optional migration.
func LoadHistory(paths Paths, migPaths *Paths) []historyEntry {
	var prompts []historyEntry
	seen := make(map[string]bool)

	var sources []string
	if migPaths != nil {
		sources = append(sources, migPaths.HistoryJSONL)
	}
	sources = append(sources, paths.HistoryJSONL)

	for _, path := range sources {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(line), &obj); err != nil {
				continue
			}
			sid := getString(obj, "sessionId")
			ts := 0.0
			if v, ok := obj["timestamp"].(float64); ok {
				ts = v
			}
			key := fmt.Sprintf("%s|%f", sid, ts)
			if seen[key] {
				continue
			}
			seen[key] = true
			prompts = append(prompts, historyEntry{
				Display:   getString(obj, "display"),
				Timestamp: ts,
				Project:   getString(obj, "project"),
				SessionID: sid,
			})
		}
		f.Close()
	}

	sort.Slice(prompts, func(i, j int) bool {
		return prompts[i].Timestamp < prompts[j].Timestamp
	})
	return prompts
}
