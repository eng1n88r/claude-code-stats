package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/AeternaLabsHQ/claude-code-stats/internal/extract"
)

func RenderActivity(data *extract.DashboardData, width int) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("87"))
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))

	barWidth := 30
	if width > 80 {
		barWidth = 40
	}

	// Hourly Distribution
	b.WriteString(title.Render("Hourly Distribution (Messages)"))
	b.WriteString("\n\n")

	maxHourly := 0
	for _, h := range data.Hourly {
		if h.Messages > maxHourly {
			maxHourly = h.Messages
		}
	}

	for _, h := range data.Hourly {
		bar := makeBar(float64(h.Messages), float64(maxHourly), barWidth)
		hourLabel := fmt.Sprintf("%02d:00", h.Hour)
		b.WriteString(fmt.Sprintf("  %s %s %s\n",
			muted.Render(hourLabel),
			cyan.Render(bar),
			fmt.Sprintf("%d", h.Messages)))
	}

	// Weekday Distribution
	b.WriteString("\n")
	b.WriteString(title.Render("Weekday Distribution"))
	b.WriteString("\n\n")

	maxWeekday := 0
	for _, w := range data.Weekday {
		if w.Messages > maxWeekday {
			maxWeekday = w.Messages
		}
	}

	for i, w := range data.Weekday {
		bar := makeBar(float64(w.Messages), float64(maxWeekday), barWidth)
		style := cyan
		if i >= 5 { // Weekend
			style = yellow
		}
		b.WriteString(fmt.Sprintf("  %-5s %s %d\n",
			w.Day,
			style.Render(bar),
			w.Messages))
	}

	// Daily Messages (last 30 days)
	b.WriteString("\n")
	b.WriteString(title.Render("Daily Messages (last 30 days)"))
	b.WriteString("\n\n")

	msgs := data.DailyMessages
	start := 0
	if len(msgs) > 30 {
		start = len(msgs) - 30
	}

	rows := make([][]string, 0)
	for i := start; i < len(msgs); i++ {
		m := msgs[i]
		rows = append(rows, []string{
			m.Date,
			fmt.Sprintf("%d", m.Messages),
			fmt.Sprintf("%d", m.Sessions),
		})
	}

	if len(rows) > 0 {
		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).Padding(0, 1)
				}
				return lipgloss.NewStyle().Padding(0, 1)
			}).
			Headers("Date", "Messages", "Sessions").
			Rows(rows...)

		b.WriteString(t.String())
	}

	return b.String()
}
