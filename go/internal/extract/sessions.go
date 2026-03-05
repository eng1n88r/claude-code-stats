package extract

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ParseSessionTranscripts parses all session JSONL transcripts from the given
// projects directories (primary + optional migration).
func ParseSessionTranscripts(paths Paths, migPaths *Paths, quiet bool) map[string]*rawSession {
	sessions := make(map[string]*rawSession)

	type source struct {
		label string
		dir   string
	}
	var sources []source
	if migPaths != nil {
		if info, err := os.Stat(migPaths.ProjectsDir); err == nil && info.IsDir() {
			sources = append(sources, source{"migration", migPaths.ProjectsDir})
		}
	}
	if info, err := os.Stat(paths.ProjectsDir); err == nil && info.IsDir() {
		sources = append(sources, source{"current", paths.ProjectsDir})
	}
	if len(sources) == 0 {
		if !quiet {
			fmt.Fprintln(os.Stderr, "  WARNING: No projects directories found")
		}
		return sessions
	}

	totalFiles := 0
	totalLines := 0

	for _, src := range sources {
		if !quiet {
			fmt.Fprintf(os.Stderr, "  Source: %s (%s)\n", src.label, src.dir)
		}

		entries, err := os.ReadDir(src.dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			projectName := entry.Name()
			projectDir := filepath.Join(src.dir, projectName)

			jsonlFiles := findJSONLFiles(projectDir)
			if len(jsonlFiles) == 0 {
				continue
			}

			for _, jf := range jsonlFiles {
				totalFiles++
				fileSessionID := strings.TrimSuffix(filepath.Base(jf), ".jsonl")

				// Skip if already parsed from migration
				if _, exists := sessions[fileSessionID]; exists && src.label == "current" {
					continue
				}

				fi, _ := os.Stat(jf)
				fileSize := int64(0)
				if fi != nil {
					fileSize = fi.Size()
				}

				totalLines += parseJSONLFile(jf, fileSessionID, projectName, fileSize, src.label, sessions)
			}
		}
	}

	if !quiet {
		migCount := 0
		curCount := 0
		for _, s := range sessions {
			if s.Source == "migration" {
				migCount++
			} else {
				curCount++
			}
		}
		fmt.Fprintf(os.Stderr, "  Parsed %d files, %d lines, %d sessions (migration: %d, current: %d)\n",
			totalFiles, totalLines, len(sessions), migCount, curCount)
	}

	return sessions
}

func findJSONLFiles(dir string) []string {
	var files []string
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files
}

func parseJSONLFile(path, fileSessionID, projectName string, fileSize int64, sourceLabel string, sessions map[string]*rawSession) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024) // 2MB max line

	lineCount := 0
	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			continue
		}

		msgType, _ := obj["type"].(string)
		sessionID := getString(obj, "sessionId")
		if sessionID == "" {
			sessionID = fileSessionID
		}

		sess, exists := sessions[sessionID]
		if !exists {
			sess = &rawSession{
				SessionID:  sessionID,
				ProjectDir: projectName,
				Models:     make(map[string]*modelAccum),
				Tools:      make(map[string]int),
				FileSize:   fileSize,
				Slug:       getString(obj, "slug"),
				Source:     sourceLabel,
			}
			if cwd := getString(obj, "cwd"); cwd != "" {
				sess.ProjectPath = cwd
			}
			sessions[sessionID] = sess
		}

		if cwd := getString(obj, "cwd"); cwd != "" && sess.ProjectPath == "" {
			sess.ProjectPath = cwd
		}
		if slug := getString(obj, "slug"); slug != "" && sess.Slug == "" {
			sess.Slug = slug
		}

		// Collect timestamps
		if ts := obj["timestamp"]; ts != nil {
			if tsMs := parseTimestamp(ts); tsMs > 0 {
				sess.Timestamps = append(sess.Timestamps, tsMs)
			}
		}

		switch msgType {
		case "user":
			sess.MessageCount++
			sess.UserMsgCount++

			if sess.FirstPrompt == "" {
				text := extractPromptText(obj)
				if text != "" &&
					!strings.HasPrefix(text, "<command") &&
					!strings.HasPrefix(text, "<local-command") &&
					!strings.HasPrefix(text, "[Request interrupted") {
					if len(text) > 200 {
						text = text[:200]
					}
					sess.FirstPrompt = text
				}
			}

		case "assistant":
			sess.MessageCount++
			sess.AssistMsgCount++

			message, _ := obj["message"].(map[string]interface{})
			if message == nil {
				continue
			}

			model := getString(message, "model")
			if model == "" {
				model = "unknown"
			}
			usage, _ := message["usage"].(map[string]interface{})

			if usage != nil && getInt(usage, "output_tokens") > 0 {
				m := sess.Models[model]
				if m == nil {
					m = &modelAccum{}
					sess.Models[model] = m
				}

				inputTokens := getInt(usage, "input_tokens")
				outputTokens := getInt(usage, "output_tokens")
				cacheRead := getInt(usage, "cache_read_input_tokens")
				cacheCreation := getInt(usage, "cache_creation_input_tokens")

				m.InputTokens += inputTokens
				m.OutputTokens += outputTokens
				m.CacheRead += cacheRead
				m.CacheCreation += cacheCreation
				m.Cost += CalcCost(model, inputTokens, outputTokens, cacheRead, cacheCreation)
				m.Calls++
			}

			// Tool usage from content blocks
			if content, ok := message["content"].([]interface{}); ok {
				for _, block := range content {
					if bm, ok := block.(map[string]interface{}); ok {
						if getString(bm, "type") == "tool_use" {
							toolName := getString(bm, "name")
							if toolName == "" {
								toolName = "unknown"
							}
							sess.Tools[toolName]++
						}
					}
				}
			}
		}
	}

	return lineCount
}

// parseTimestamp handles both ISO 8601 strings and Unix millisecond numbers.
func parseTimestamp(v interface{}) int64 {
	switch ts := v.(type) {
	case string:
		// Try ISO 8601
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			t, err = time.Parse(time.RFC3339Nano, ts)
		}
		if err != nil {
			// Try with Z replacement
			t, err = time.Parse(time.RFC3339, strings.Replace(ts, "Z", "+00:00", 1))
		}
		if err != nil {
			return 0
		}
		return t.UnixMilli()
	case float64:
		return int64(ts)
	case json.Number:
		if n, err := ts.Int64(); err == nil {
			return n
		}
	}
	return 0
}

func extractPromptText(obj map[string]interface{}) string {
	message, _ := obj["message"].(map[string]interface{})
	if message == nil {
		return ""
	}
	content := message["content"]
	switch c := content.(type) {
	case string:
		return c
	case []interface{}:
		for _, block := range c {
			if bm, ok := block.(map[string]interface{}); ok {
				if getString(bm, "type") == "text" {
					return getString(bm, "text")
				}
			}
		}
	}
	return ""
}

func getString(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func getInt(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case json.Number:
		n, _ := v.Int64()
		return int(n)
	}
	return 0
}
