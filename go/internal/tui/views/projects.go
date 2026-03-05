package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/AeternaLabsHQ/claude-code-stats/internal/extract"
)

func RenderProjects(data *extract.DashboardData, limit int) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true)

	b.WriteString(title.Render("Projects"))
	b.WriteString("\n\n")

	projects := data.Projects
	showing := len(projects)
	if limit > 0 && limit < showing {
		showing = limit
	}

	rows := make([][]string, 0, showing+1)
	var totalSessions, totalMessages, totalOutput int
	var totalCost float64
	var totalSize float64

	for i := 0; i < showing; i++ {
		p := projects[i]
		rows = append(rows, []string{
			p.Name,
			fmt.Sprintf("%d", p.Sessions),
			fmt.Sprintf("%d", p.Messages),
			fmtCost(p.Cost),
			fmtTokens(p.OutputTokens),
			fmt.Sprintf("%.1f", p.FileSizeMB),
		})
	}

	for _, p := range projects {
		totalSessions += p.Sessions
		totalMessages += p.Messages
		totalCost += p.Cost
		totalOutput += p.OutputTokens
		totalSize += p.FileSizeMB
	}

	// Totals row
	rows = append(rows, []string{
		fmt.Sprintf("Total (%d)", len(projects)),
		fmt.Sprintf("%d", totalSessions),
		fmt.Sprintf("%d", totalMessages),
		fmtCost(totalCost),
		fmtTokens(totalOutput),
		fmt.Sprintf("%.1f", totalSize),
	})

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return s.Bold(true).Foreground(lipgloss.Color("99"))
			}
			if row == len(rows)-1 { // totals row
				return s.Bold(true).Foreground(lipgloss.Color("255"))
			}
			return s
		}).
		Headers("Project", "Sessions", "Messages", "API Value", "Output Tokens", "File Size (MB)").
		Rows(rows...)

	b.WriteString(t.String())

	return b.String()
}
