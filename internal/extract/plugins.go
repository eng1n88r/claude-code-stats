package extract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// LoadPlugins loads plugin data from primary + optional migration.
func LoadPlugins(paths Paths, migPaths *Paths) PluginsData {
	result := PluginsData{
		MarketplaceStats: make(map[string]int),
	}
	seenPlugins := make(map[string]bool)

	var sources []string
	if migPaths != nil {
		sources = append(sources, migPaths.ClaudeDir)
	}
	sources = append(sources, paths.ClaudeDir)

	for _, claudeDir := range sources {
		pluginsDir := filepath.Join(claudeDir, "plugins")

		// Installed plugins
		installedFile := filepath.Join(pluginsDir, "installed_plugins.json")
		data, err := os.ReadFile(installedFile)
		if err == nil {
			var raw struct {
				Plugins map[string][]struct {
					Version     string `json:"version"`
					InstalledAt string `json:"installedAt"`
					LastUpdated string `json:"lastUpdated"`
				} `json:"plugins"`
			}
			if json.Unmarshal(data, &raw) == nil {
				for name, versions := range raw.Plugins {
					if len(versions) == 0 || seenPlugins[name] {
						continue
					}
					seenPlugins[name] = true
					v := versions[0]
					shortName := name
					marketplace := ""
					if idx := strings.Index(name, "@"); idx >= 0 {
						shortName = name[:idx]
						marketplace = name[idx+1:]
					}
					result.Installed = append(result.Installed, PluginInfo{
						Name:        name,
						ShortName:   shortName,
						Marketplace: marketplace,
						Version:     v.Version,
						InstalledAt: v.InstalledAt,
						LastUpdated: v.LastUpdated,
					})
				}
			}
		}

		// Marketplace install counts
		countsFile := filepath.Join(pluginsDir, "install-counts-cache.json")
		data, err = os.ReadFile(countsFile)
		if err == nil {
			var raw struct {
				Counts []struct {
					Plugin         string `json:"plugin"`
					UniqueInstalls int    `json:"unique_installs"`
				} `json:"counts"`
			}
			if json.Unmarshal(data, &raw) == nil {
				for _, c := range raw.Counts {
					result.MarketplaceStats[c.Plugin] = c.UniqueInstalls
				}
			}
		}
	}

	// Settings from primary only
	settingsFile := filepath.Join(paths.ClaudeDir, "settings.json")
	data, err := os.ReadFile(settingsFile)
	if err == nil {
		var raw struct {
			Permissions struct {
				DefaultMode string `json:"defaultMode"`
			} `json:"permissions"`
			AutoUpdatesChannel string          `json:"autoUpdatesChannel"`
			EnabledPlugins     map[string]bool `json:"enabledPlugins"`
		}
		if json.Unmarshal(data, &raw) == nil {
			result.Settings = PluginSettings{
				PermissionMode: raw.Permissions.DefaultMode,
				AutoUpdates:    raw.AutoUpdatesChannel,
				EnabledPlugins: raw.EnabledPlugins,
			}
		}
	}

	return result
}
