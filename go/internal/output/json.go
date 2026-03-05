package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AeternaLabsHQ/claude-code-stats/internal/extract"
)

// WriteJSON writes the dashboard data to a JSON file.
func WriteJSON(data *extract.DashboardData, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	outPath := filepath.Join(outputDir, "dashboard_data.json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if err := os.WriteFile(outPath, jsonData, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}

	fmt.Fprintf(os.Stderr, "  Data written to: %s\n", outPath)
	return nil
}

// DumpJSON writes the dashboard data as JSON to stdout.
func DumpJSON(data *extract.DashboardData) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
