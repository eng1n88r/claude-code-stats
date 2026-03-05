package extract

import (
	"math"
	"os"
	"path/filepath"
	"sort"
)

// CalcStorage calculates storage breakdown for ~/.claude/ + optional migration.
func CalcStorage(paths Paths, migPaths *Paths) StorageData {
	breakdown := make(map[string]int64)
	var total int64

	entries, err := os.ReadDir(paths.ClaudeDir)
	if err == nil {
		for _, entry := range entries {
			fullPath := filepath.Join(paths.ClaudeDir, entry.Name())
			if entry.IsDir() {
				sz := dirSize(fullPath)
				breakdown[entry.Name()+"/"] = sz
				total += sz
			} else {
				fi, err := entry.Info()
				if err == nil {
					breakdown[entry.Name()] = fi.Size()
					total += fi.Size()
				}
			}
		}
	}

	// Migration backup as single entry
	if migPaths != nil {
		if info, err := os.Stat(migPaths.ClaudeDir); err == nil && info.IsDir() {
			sz := dirSize(migPaths.ClaudeDir)
			// Also include the dot-claude JSON
			if fi, err := os.Stat(migPaths.DotClaudeJSON); err == nil {
				sz += fi.Size()
			}
			if sz > 0 {
				breakdown["_migration-backup/"] = sz
				total += sz
			}
		}
	}

	// Sort by size descending
	type kv struct {
		name string
		size int64
	}
	var sorted_ []kv
	for k, v := range breakdown {
		if v > 0 {
			sorted_ = append(sorted_, kv{k, v})
		}
	}
	sort.Slice(sorted_, func(i, j int) bool {
		return sorted_[i].size > sorted_[j].size
	})

	var items []StorageItem
	for _, kv := range sorted_ {
		items = append(items, StorageItem{
			Name:   kv.name,
			SizeMB: math.Round(float64(kv.size)/1048576*100) / 100,
		})
	}

	return StorageData{
		TotalMB: math.Round(float64(total)/1048576*10) / 10,
		Items:   items,
	}
}

// LoadFileHistoryStats counts files in file-history directories.
func LoadFileHistoryStats(paths Paths, migPaths *Paths) FileHistoryData {
	var totalFiles, sessions int
	var totalSize int64
	seenSessions := make(map[string]bool)

	var sources []string
	if migPaths != nil {
		sources = append(sources, migPaths.ClaudeDir)
	}
	sources = append(sources, paths.ClaudeDir)

	for _, claudeDir := range sources {
		fhDir := filepath.Join(claudeDir, "file-history")
		entries, err := os.ReadDir(fhDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() || seenSessions[entry.Name()] {
				continue
			}
			seenSessions[entry.Name()] = true
			sessions++

			sessDir := filepath.Join(fhDir, entry.Name())
			files, err := os.ReadDir(sessDir)
			if err != nil {
				continue
			}
			for _, f := range files {
				if f.IsDir() {
					continue
				}
				totalFiles++
				if fi, err := f.Info(); err == nil {
					totalSize += fi.Size()
				}
			}
		}
	}

	return FileHistoryData{
		TotalFiles:    totalFiles,
		TotalSessions: sessions,
		TotalSizeMB:   math.Round(float64(totalSize)/1048576*10) / 10,
	}
}

func dirSize(path string) int64 {
	var size int64
	_ = filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		fi, err := d.Info()
		if err == nil {
			size += fi.Size()
		}
		return nil
	})
	return size
}
