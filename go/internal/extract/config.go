package extract

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed locales/*.json
var localeFS embed.FS

// LoadConfig reads config.json from the given path.
// Returns an error if the file doesn't exist.
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config not found: %s — copy config.example.json and adjust", configPath)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config.json: %w", err)
	}
	return &cfg, nil
}

// FindConfig looks for config.json in standard locations:
// 1. Explicit path (if provided)
// 2. CWD
// 3. ~/.config/claude-dashboard/
// Returns the path found, or error if none exists.
func FindConfig(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err == nil {
			return explicit, nil
		}
		return "", fmt.Errorf("config not found at %s", explicit)
	}

	// CWD
	cwd, _ := os.Getwd()
	p := filepath.Join(cwd, "config.json")
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}

	// ~/.config/claude-dashboard/
	home, err := os.UserHomeDir()
	if err == nil {
		p = filepath.Join(home, ".config", "claude-dashboard", "config.json")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("config.json not found in CWD or ~/.config/claude-dashboard/")
}

// LoadLocale loads a locale JSON file. First tries the embedded FS,
// then falls back to a locales/ directory relative to the executable.
// Returns the raw JSON bytes (passed through to DashboardData.Locale).
func LoadLocale(lang string) (json.RawMessage, error) {
	fname := lang + ".json"

	// Try embedded
	data, err := localeFS.ReadFile("locales/" + fname)
	if err == nil {
		return json.RawMessage(data), nil
	}

	// Fallback: try alongside executable
	exe, _ := os.Executable()
	if exe != "" {
		p := filepath.Join(filepath.Dir(exe), "locales", fname)
		data, err = os.ReadFile(p)
		if err == nil {
			return json.RawMessage(data), nil
		}
	}

	// Fallback to en.json
	if lang != "en" {
		fmt.Fprintf(os.Stderr, "WARNING: Locale '%s' not found, falling back to 'en'\n", lang)
		return LoadLocale("en")
	}

	return nil, fmt.Errorf("locale file not found: %s", fname)
}

// ClaudeDir returns the path to ~/.claude/
func ClaudeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

// Paths holds all resolved file/directory paths for data extraction.
type Paths struct {
	ClaudeDir      string
	ProjectsDir    string
	DotClaudeJSON  string
	StatsCache     string
	HistoryJSONL   string
	OutputDir      string
	DashboardData  string
	DashboardHTML  string
	TemplateHTML   string
}

// ResolvePaths builds the primary Paths from the Claude home directory.
func ResolvePaths(baseDir string) Paths {
	home, _ := os.UserHomeDir()
	claudeDir := filepath.Join(home, ".claude")
	if baseDir != "" {
		claudeDir = baseDir
	}
	dotClaude := filepath.Join(home, ".claude.json")

	return Paths{
		ClaudeDir:     claudeDir,
		ProjectsDir:   filepath.Join(claudeDir, "projects"),
		DotClaudeJSON: dotClaude,
		StatsCache:    filepath.Join(claudeDir, "stats-cache.json"),
		HistoryJSONL:  filepath.Join(claudeDir, "history.jsonl"),
	}
}

// MigrationPaths resolves paths for the migration backup source.
func MigrationPaths(cfg *Config) *Paths {
	mig := cfg.Migration
	if !mig.Enabled || mig.Dir == nil {
		return nil
	}
	dir := os.ExpandEnv(*mig.Dir)
	// Expand ~ manually
	if len(dir) > 0 && dir[0] == '~' {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, dir[1:])
	}

	claudeDirName := mig.ClaudeDirName
	if claudeDirName == "" {
		claudeDirName = ".claude-windows"
	}
	dotClaudeName := mig.DotClaudeJSONName
	if dotClaudeName == "" {
		dotClaudeName = ".claude-windows.json"
	}

	claudeDir := filepath.Join(dir, claudeDirName)
	return &Paths{
		ClaudeDir:     claudeDir,
		ProjectsDir:   filepath.Join(claudeDir, "projects"),
		DotClaudeJSON: filepath.Join(dir, dotClaudeName),
		StatsCache:    filepath.Join(claudeDir, "stats-cache.json"),
		HistoryJSONL:  filepath.Join(claudeDir, "history.jsonl"),
	}
}
