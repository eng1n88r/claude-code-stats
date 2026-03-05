package tui

import (
	"charm.land/lipgloss/v2"
)

var (
	// Colors
	colorPrimary   = lipgloss.Color("99")  // purple
	colorAccent    = lipgloss.Color("212") // pink
	colorGreen     = lipgloss.Color("82")
	colorRed       = lipgloss.Color("196")
	colorOrange    = lipgloss.Color("208")
	colorCyan      = lipgloss.Color("87")
	colorMagenta   = lipgloss.Color("213")
	colorYellow    = lipgloss.Color("220")
	colorMuted     = lipgloss.Color("243")
	colorDim       = lipgloss.Color("240")
	colorBright    = lipgloss.Color("255")
	colorBg        = lipgloss.Color("236")
	colorTabActive = lipgloss.Color("99")
	colorTabBg     = lipgloss.Color("238")

	// Tab styles
	tabStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 2)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(colorBright).
			Background(colorTabActive).
			Bold(true).
			Padding(0, 2)

	tabGap = lipgloss.NewStyle().
		Foreground(colorDim).
		SetString(" │ ")

	// KPI styles
	kpiLabel = lipgloss.NewStyle().
			Foreground(colorMuted)

	kpiValue = lipgloss.NewStyle().
			Bold(true)

	kpiSep = lipgloss.NewStyle().
		Foreground(colorDim).
		SetString("  │  ")

	// Content styles
	sectionTitle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginBottom(1)

	dimText = lipgloss.NewStyle().
		Foreground(colorMuted)

	brightText = lipgloss.NewStyle().
			Foreground(colorBright)

	greenText = lipgloss.NewStyle().
			Foreground(colorGreen)

	redText = lipgloss.NewStyle().
		Foreground(colorRed)

	orangeText = lipgloss.NewStyle().
			Foreground(colorOrange)

	cyanText = lipgloss.NewStyle().
		Foreground(colorCyan)

	magentaText = lipgloss.NewStyle().
			Foreground(colorMagenta)

	yellowText = lipgloss.NewStyle().
			Foreground(colorYellow)

	// Help style
	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			MarginTop(1)
)

// Bar characters for charts
const (
	barFull  = "█"
	barEmpty = "░"
)

func makeBar(val, maxVal float64, width int) string {
	if maxVal <= 0 {
		return ""
	}
	filled := int(val / maxVal * float64(width))
	if filled > width {
		filled = width
	}
	bar := ""
	for i := 0; i < filled; i++ {
		bar += barFull
	}
	for i := filled; i < width; i++ {
		bar += barEmpty
	}
	return bar
}
