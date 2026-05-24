package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewPicker() string {
	if len(m.pickerItems) == 0 {
		return lipgloss.NewStyle().Italic(true).Render("(nothing to select)")
	}

	title := "Select condition type:"
	if m.pickerKind == FilterByAdapter {
		title = "Select adapter:"
	}

	faint := lipgloss.NewStyle().Faint(true)
	selected := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	if m.opts.NoColor {
		selected = lipgloss.NewStyle().Bold(true)
	}

	visible := m.mainPanelHeight()
	start := 0
	if m.pickerSelected >= visible {
		start = m.pickerSelected - visible + 1
	}
	end := start + visible
	if end > len(m.pickerItems) {
		end = len(m.pickerItems)
	}

	var lines []string
	lines = append(lines, faint.Render(title), "")
	for i := start; i < end; i++ {
		item := m.pickerItems[i]
		prefix := "  "
		line := item
		if i == m.pickerSelected {
			prefix = "> "
			line = selected.Render(item)
		}
		lines = append(lines, prefix+line)
	}
	if len(m.pickerItems) > visible {
		lines = append(lines, "", faint.Render(fmt.Sprintf("%d/%d", m.pickerSelected+1, len(m.pickerItems))))
	}
	return strings.Join(lines, "\n")
}
