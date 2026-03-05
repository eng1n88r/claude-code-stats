package extract

import (
	"testing"
	"time"
)

func TestClampedDate(t *testing.T) {
	tests := []struct {
		name    string
		year    int
		month   time.Month
		day     int
		wantDay int
	}{
		{"jan 31 stays 31", 2026, time.January, 31, 31},
		{"feb 31 clamps to 28", 2026, time.February, 31, 28},
		{"feb 29 in leap year", 2024, time.February, 29, 29},
		{"feb 29 in non-leap clamps to 28", 2026, time.February, 29, 28},
		{"apr 31 clamps to 30", 2026, time.April, 31, 30},
		{"jun 30 stays 30", 2026, time.June, 30, 30},
		{"normal day", 2026, time.March, 15, 15},
		{"day 1", 2026, time.January, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampedDate(tt.year, tt.month, tt.day)
			if got.Day() != tt.wantDay {
				t.Errorf("clampedDate(%d, %v, %d).Day() = %d, want %d",
					tt.year, tt.month, tt.day, got.Day(), tt.wantDay)
			}
			if got.Month() != tt.month {
				t.Errorf("clampedDate(%d, %v, %d).Month() = %v, want %v",
					tt.year, tt.month, tt.day, got.Month(), tt.month)
			}
		})
	}
}

func TestBuildPlanAnalysis(t *testing.T) {
	end := "2026-02-28"
	planHistory := []PlanConfig{
		{
			Plan:       "Max",
			Start:      "2026-01-01",
			End:        &end,
			CostEUR:    ptr(87.61),
			CostUSD:    93.00,
			BillingDay: 1,
		},
	}
	dailyCosts := []map[string]any{
		{"date": "2026-01-15", "total": 50.0},
		{"date": "2026-02-10", "total": 30.0},
		{"date": "2026-03-01", "total": 99.0}, // outside period
	}
	sessions := []SessionOutput{
		{Date: "2026-01-15", Messages: 100},
		{Date: "2026-02-10", Messages: 50},
		{Date: "2026-03-01", Messages: 999}, // outside period
	}

	result := BuildPlanAnalysis(planHistory, dailyCosts, sessions)

	if len(result.Periods) != 1 {
		t.Fatalf("expected 1 period, got %d", len(result.Periods))
	}
	p := result.Periods[0]

	if p.APICost != 80.0 {
		t.Errorf("expected API cost 80.0, got %f", p.APICost)
	}
	if p.Sessions != 2 {
		t.Errorf("expected 2 sessions, got %d", p.Sessions)
	}
	if p.Messages != 150 {
		t.Errorf("expected 150 messages, got %d", p.Messages)
	}
	if p.DaysActive != 2 {
		t.Errorf("expected 2 active days, got %d", p.DaysActive)
	}
	// Savings: 80.0 - 93.0 = -13.0
	if p.Savings != -13.0 {
		t.Errorf("expected savings -13.0, got %f", p.Savings)
	}
}

func ptr(v float64) *float64 { return &v }

func TestBuildPlanAnalysis_EmptyHistory(t *testing.T) {
	result := BuildPlanAnalysis(nil, nil, nil)
	if len(result.Periods) != 0 {
		t.Errorf("expected 0 periods, got %d", len(result.Periods))
	}
	if result.CurrentBilling != nil {
		t.Error("expected nil current billing with no plan history")
	}
}
