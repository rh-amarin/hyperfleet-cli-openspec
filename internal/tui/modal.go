package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	modalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("203")).
			Padding(1, 2)

	modalTitleStyle = lipgloss.NewStyle().Bold(true)
	modalHintStyle  = lipgloss.NewStyle().Faint(true)
)

func (m Model) renderDeleteModal() string {
	title := "Delete"
	prompt := fmt.Sprintf("Delete %s?", m.deleteTargetLabel())
	if m.deleteForce {
		title = "Force-delete"
		prompt = fmt.Sprintf("Force-delete %s?", m.deleteTargetLabel())
	}

	body := strings.Join([]string{
		modalTitleStyle.Render(title),
		"",
		prompt,
		"",
		modalHintStyle.Render("y  confirm"),
		modalHintStyle.Render("n  cancel"),
	}, "\n")

	if m.opts.NoColor {
		return modalBoxStyle.Copy().BorderForeground(lipgloss.NoColor{}).Render(body)
	}
	return modalBoxStyle.Render(body)
}

// overlayModal centers modal lines over a fixed-size base view.
func overlayModal(base, modal string, width, height int) string {
	if width < 1 {
		width = 80
	}
	if height < 1 {
		height = 24
	}

	baseLines := strings.Split(base, "\n")
	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}
	if len(baseLines) > height {
		baseLines = baseLines[:height]
	}

	modalLines := strings.Split(modal, "\n")
	startY := (height - len(modalLines)) / 2
	if startY < 0 {
		startY = 0
	}

	for i, ml := range modalLines {
		y := startY + i
		if y >= height {
			break
		}
		baseLines[y] = centerLine(ml, width)
	}
	return strings.Join(baseLines, "\n")
}

func centerLine(line string, width int) string {
	pad := (width - lipgloss.Width(line)) / 2
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad) + line
}
