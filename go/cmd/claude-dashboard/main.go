package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AeternaLabsHQ/claude-code-stats/internal/extract"
	"github.com/AeternaLabsHQ/claude-code-stats/internal/output"
	"github.com/AeternaLabsHQ/claude-code-stats/internal/tui"
	"github.com/spf13/cobra"
)

var version = "dev"

var (
	flagJSON      bool
	flagAll       bool
	flagSection   string
	flagLimit     int
	flagNoRefresh bool
	flagQuiet     bool
	flagConfig    string
	flagOutput    string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "claude-dashboard",
		Short: "Claude Code usage statistics dashboard",
		Long:  "Extract stats from ~/.claude/ and display an interactive TUI dashboard.",
		RunE:  runRoot,
	}

	rootCmd.Flags().BoolVar(&flagJSON, "json", false, "Dump JSON to stdout")
	rootCmd.Flags().BoolVar(&flagAll, "all", false, "Print all sections non-interactively")
	rootCmd.Flags().StringVar(&flagSection, "section", "", "Print specific sections (comma-separated)")
	rootCmd.Flags().IntVar(&flagLimit, "limit", 20, "Limit rows in tables")
	rootCmd.Flags().BoolVar(&flagNoRefresh, "no-refresh", false, "Skip extraction, use existing data")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "Suppress progress output")
	rootCmd.PersistentFlags().StringVar(&flagConfig, "config", "", "Path to config.json")
	rootCmd.PersistentFlags().StringVar(&flagOutput, "output", "", "Output directory (default: ./public)")

	// extract subcommand
	extractCmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract data only (no TUI), write JSON + HTML",
		RunE:  runExtract,
	}
	rootCmd.AddCommand(extractCmd)

	// version subcommand
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("claude-dashboard %s\n", version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func loadConfigAndExtract() (*extract.DashboardData, error) {
	configPath, err := extract.FindConfig(flagConfig)
	if err != nil {
		return nil, err
	}
	cfg, err := extract.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return extract.Run(cfg, flagQuiet)
}

func outputDir() string {
	if flagOutput != "" {
		return flagOutput
	}
	return filepath.Join(".", "public")
}

func dataFilePath() string {
	return filepath.Join(outputDir(), "dashboard_data.json")
}

// loadExistingData reads dashboard_data.json from disk.
func loadExistingData() (*extract.DashboardData, error) {
	raw, err := os.ReadFile(dataFilePath())
	if err != nil {
		return nil, fmt.Errorf("no existing data: %w", err)
	}
	var data extract.DashboardData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("invalid dashboard_data.json: %w", err)
	}
	return &data, nil
}

// isDataStale returns true if the data file is missing or older than 10 minutes.
func isDataStale() bool {
	fi, err := os.Stat(dataFilePath())
	if err != nil {
		return true
	}
	return time.Since(fi.ModTime()) > 10*time.Minute
}

// getData either loads existing data (--no-refresh) or extracts fresh data.
// Auto-refresh: re-extract if data is >10 minutes old (unless --no-refresh).
func getData() (*extract.DashboardData, error) {
	if flagNoRefresh {
		return loadExistingData()
	}
	if !isDataStale() {
		// Data is fresh, load from disk
		data, err := loadExistingData()
		if err == nil {
			return data, nil
		}
		// Fall through to extract if load fails
	}
	return loadConfigAndExtract()
}

func runRoot(cmd *cobra.Command, args []string) error {
	if flagJSON {
		data, err := loadConfigAndExtract()
		if err != nil {
			return err
		}
		return output.DumpJSON(data)
	}

	// For --all / --section, extract synchronously
	if flagAll || flagSection != "" {
		data, err := getData()
		if err != nil {
			return err
		}
		if !flagNoRefresh {
			_ = output.WriteJSON(data, outputDir())
		}
		return printSections(data)
	}

	// For TUI: if --no-refresh or data is fresh, load and launch immediately
	if flagNoRefresh || !isDataStale() {
		data, err := loadExistingData()
		if err == nil {
			return tui.Run(data, flagLimit)
		}
		// Fall through to async extraction
	}

	// Launch TUI with spinner while extracting in background
	return tui.RunWithExtraction(flagLimit, func() (*extract.DashboardData, error) {
		data, err := loadConfigAndExtract()
		if err != nil {
			return nil, err
		}
		_ = output.WriteJSON(data, outputDir())
		return data, nil
	})
}

func printSections(data *extract.DashboardData) error {
	validSections := map[string]func(*extract.DashboardData, int) string{
		"overview": func(d *extract.DashboardData, w int) string { return tui.RenderOverviewText(d) },
		"tokens":   func(d *extract.DashboardData, w int) string { return tui.RenderTokensText(d, w) },
		"activity": func(d *extract.DashboardData, w int) string { return tui.RenderActivityText(d, w) },
		"projects": func(d *extract.DashboardData, w int) string { return tui.RenderProjectsText(d, flagLimit) },
		"sessions": func(d *extract.DashboardData, w int) string { return tui.RenderSessionsText(d, flagLimit) },
		"plan":     func(d *extract.DashboardData, w int) string { return tui.RenderPlanText(d) },
		"insights": func(d *extract.DashboardData, w int) string { return tui.RenderInsightsText(d) },
	}

	width := 120 // default width for non-interactive

	if flagAll {
		// Print KPI header
		fmt.Println(tui.RenderKPIText(data))
		fmt.Println()
		for _, name := range []string{"overview", "tokens", "activity", "projects", "sessions", "plan", "insights"} {
			fmt.Println(validSections[name](data, width))
			fmt.Println()
		}
		return nil
	}

	// Specific sections
	fmt.Println(tui.RenderKPIText(data))
	fmt.Println()
	for _, s := range strings.Split(flagSection, ",") {
		s = strings.TrimSpace(s)
		if fn, ok := validSections[s]; ok {
			fmt.Println(fn(data, width))
			fmt.Println()
		} else {
			fmt.Fprintf(os.Stderr, "Unknown section: %s (valid: overview, tokens, activity, projects, sessions, plan, insights)\n", s)
		}
	}
	return nil
}

func runExtract(cmd *cobra.Command, args []string) error {
	data, err := loadConfigAndExtract()
	if err != nil {
		return err
	}
	return output.WriteJSON(data, outputDir())
}
