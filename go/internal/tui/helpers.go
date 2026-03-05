package tui

import (
	"fmt"
	"strings"
)

func fmtCost(v float64) string {
	return fmt.Sprintf("$%.2f", v)
}

func fmtTokens(v int) string {
	switch {
	case v >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(v)/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("%.1fK", float64(v)/1_000)
	default:
		return fmt.Sprintf("%d", v)
	}
}

func fmtPct(v, total float64) string {
	if total <= 0 {
		return "0%"
	}
	return fmt.Sprintf("%.0f%%", v/total*100)
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
