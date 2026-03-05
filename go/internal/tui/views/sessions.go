package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/AeternaLabsHQ/claude-code-stats/internal/extract"
)

func RenderSessions(data *extract.DashboardData, limit int) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	b.WriteString(title.Render("Sessions"))
	b.WriteString("\n\n")

	sessions := data.Sessions
	// Reverse to show newest first
	reversed := make([]extract.SessionOutput, len(sessions))
	for i, s := range sessions {
		reversed[len(sessions)-1-i] = s
	}

	showing := len(reversed)
	if limit > 0 && limit < showing {
		showing = limit
	}

	rows := make([][]string, 0, showing)
	for i := 0; i < showing; i++ {
		s := reversed[i]
		prompt := s.FirstPrompt
		if len(prompt) > 60 {
			prompt = prompt[:57] + "..."
		}
		prompt = strings.ReplaceAll(prompt, "\n", " ")

		rows = append(rows, []string{
			s.Date,
			s.Project,
			fmt.Sprintf("%.0fm", s.DurationMin),
			fmtCost(s.Cost),
			fmt.Sprintf("%d", s.Messages),
			s.PrimaryModel,
			prompt,
		})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return s.Bold(true).Foreground(lipgloss.Color("99"))
			}
			return s
		}).
		Headers("Date", "Project", "Duration", "Cost", "Msgs", "Model", "First Prompt").
		Rows(rows...)

	b.WriteString(t.String())

	if showing < len(reversed) {
		b.WriteString("\n")
		b.WriteString(muted.Render(fmt.Sprintf("  ... and %d more sessions", len(reversed)-showing)))
	}

	return b.String()
}
