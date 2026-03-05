package tui

import "testing"

func TestFmtCost(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "$0.00"},
		{18.59, "$18.59"},
		{1234.5, "$1234.50"},
		{0.001, "$0.00"},
	}
	for _, tt := range tests {
		got := fmtCost(tt.input)
		if got != tt.want {
			t.Errorf("fmtCost(%f) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFmtTokens(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{500, "500"},
		{999, "999"},
		{1000, "1.0K"},
		{6300, "6.3K"},
		{999999, "1000.0K"},
		{1000000, "1.0M"},
		{1200000, "1.2M"},
		{45_600_000, "45.6M"},
	}
	for _, tt := range tests {
		got := fmtTokens(tt.input)
		if got != tt.want {
			t.Errorf("fmtTokens(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFmtPct(t *testing.T) {
	tests := []struct {
		v, total float64
		want     string
	}{
		{50, 100, "50%"},
		{1, 3, "33%"},
		{0, 0, "0%"},
		{100, 100, "100%"},
	}
	for _, tt := range tests {
		got := fmtPct(tt.v, tt.total)
		if got != tt.want {
			t.Errorf("fmtPct(%f, %f) = %q, want %q", tt.v, tt.total, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is..."},
		{"exact10ch!", 10, "exact10ch!"},
		{"multi\nline\ntext", 20, "multi line text"},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}
