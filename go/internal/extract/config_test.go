package extract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "..", "testdata", "test_config.json")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("testdata/test_config.json not found, skipping")
	}

	cfg, err := LoadConfig(fixturePath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Language != "en" {
		t.Errorf("expected language 'en', got %q", cfg.Language)
	}
	if len(cfg.PlanHistory) != 1 {
		t.Fatalf("expected 1 plan history entry, got %d", len(cfg.PlanHistory))
	}
	ph := cfg.PlanHistory[0]
	if ph.Plan != "Max" {
		t.Errorf("expected plan 'Max', got %q", ph.Plan)
	}
	if ph.CostUSD != 93.00 {
		t.Errorf("expected cost_usd 93.00, got %f", ph.CostUSD)
	}
	if ph.BillingDay != 1 {
		t.Errorf("expected billing_day 1, got %d", ph.BillingDay)
	}
	if ph.End != nil {
		t.Errorf("expected end nil, got %v", ph.End)
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.json")
	if err == nil {
		t.Error("expected error for nonexistent config")
	}
}

func TestLoadLocale(t *testing.T) {
	locale, err := LoadLocale("en")
	if err != nil {
		t.Fatalf("LoadLocale(en) failed: %v", err)
	}
	if len(locale) == 0 {
		t.Error("expected non-empty locale data")
	}
}

func TestLoadLocale_Fallback(t *testing.T) {
	// Should fall back to "en" for unknown language
	locale, err := LoadLocale("xx")
	if err != nil {
		t.Fatalf("LoadLocale(xx) failed: %v", err)
	}
	if len(locale) == 0 {
		t.Error("expected non-empty fallback locale data")
	}
}

func TestFindConfig_Explicit(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "..", "testdata", "test_config.json")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("testdata/test_config.json not found, skipping")
	}

	path, err := FindConfig(fixturePath)
	if err != nil {
		t.Fatalf("FindConfig with explicit path failed: %v", err)
	}
	if path != fixturePath {
		t.Errorf("expected %q, got %q", fixturePath, path)
	}
}

func TestFindConfig_ExplicitNotFound(t *testing.T) {
	_, err := FindConfig("/nonexistent/config.json")
	if err == nil {
		t.Error("expected error for nonexistent explicit config")
	}
}
