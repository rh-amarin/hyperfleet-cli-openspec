package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
)

func (m Model) headerLineCount() int {
	lines := 3
	if m.detailOpen || m.viewMode == ViewPicker || m.viewMode == ViewFilter {
		lines = 4
	}
	if m.context.AutoPortForward {
		lines++
	}
	return lines
}

func (m Model) renderHeader() string {
	lines := []string{m.headerViewLine()}
	if m.detailOpen || m.viewMode == ViewPicker || m.viewMode == ViewFilter {
		if rl := m.headerResourceLine(); rl != "" {
			lines = append(lines, rl)
		}
	}
	if autoLine := m.headerAutoPortForwardLine(); autoLine != "" {
		lines = append(lines, autoLine)
	}
	lines = append(lines, m.headerConfigLine(), m.headerPortForwardLine())
	return strings.Join(lines, "\n")
}

func (m Model) renderHeaderSeparator() string {
	width := m.width
	if width < 1 {
		width = 80
	}
	line := strings.Repeat("─", width)
	if m.opts.NoColor {
		return line
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(line)
}

func (m Model) headerViewLine() string {
	bold := lipgloss.NewStyle().Bold(true)
	faint := lipgloss.NewStyle().Faint(true)

	var view string
	switch {
	case m.detailOpen:
		view = fmt.Sprintf("View: Describe [%s]", m.detailFormatLabel())
	case m.viewMode == ViewPicker:
		if m.pickerKind == FilterByAdapter {
			view = "View: Select adapter"
		} else {
			view = "View: Select condition type"
		}
	case m.viewMode == ViewFilter:
		if m.filterKind == FilterByAdapter {
			view = fmt.Sprintf("View: Adapter %s", m.filterKey)
		} else {
			view = fmt.Sprintf("View: Condition %s", m.filterKey)
		}
	default:
		view = "View: Clusters/Nodepools"
	}

	if m.statusMsg != "" {
		style := bold
		if strings.HasPrefix(m.statusMsg, "[ERROR]") {
			style = style.Foreground(lipgloss.Color("9"))
		}
		return style.Render(m.statusMsg)
	}

	right := fmt.Sprintf("↻ %ds  %s", m.secsLeft, output.SpinnerFrame(m.tick))
	if m.detailOpen || m.viewMode == ViewPicker || m.viewMode == ViewFilter {
		return bold.Render(view)
	}
	return bold.Render(view) + faint.Render("  "+right)
}

func (m Model) headerResourceLine() string {
	cl, np, _, ok := m.snapshot.SelectedResource(m.selected)
	if !ok {
		return ""
	}
	bold := lipgloss.NewStyle().Bold(true)
	faint := lipgloss.NewStyle().Faint(true)
	if np != nil {
		return bold.Render("nodepool:") + " " + np.Name + "  " + faint.Render(np.ID)
	}
	return bold.Render("cluster:") + " " + cl.Name + "  " + faint.Render(cl.ID)
}

func (m Model) headerConfigLine() string {
	ctx := m.context
	faint := lipgloss.NewStyle().Faint(true)
	label := faint.Render
	value := lipgloss.NewStyle().Render

	env := ctx.ActiveEnv
	if env == "" {
		env = "(none)"
	}
	api := ctx.APIURL
	if api == "" {
		api = "(unset)"
	}
	k8s := ctx.KubeContext
	if k8s == "" {
		k8s = "(default)"
	}

	return strings.Join([]string{
		label("env:") + " " + value(env),
		label("api:") + " " + value(api),
		label("k8s:") + " " + value(k8s),
	}, "  ")
}

const autoPortForwardNotice = "using auto port-forward ⚠️"

func (m Model) headerAutoPortForwardLine() string {
	if !m.context.AutoPortForward {
		return ""
	}
	style := lipgloss.NewStyle().Bold(true)
	if !m.opts.NoColor {
		style = style.Foreground(lipgloss.Color("11"))
	}
	return style.Render(autoPortForwardNotice)
}

func (m Model) headerPortForwardLine() string {
	ctx := m.context
	faint := lipgloss.NewStyle().Faint(true)
	parts := []string{faint.Render("pf:")}

	if len(ctx.PortForwards) == 0 {
		if !ctx.AutoPortForward {
			parts = append(parts, faint.Render("(none configured)"))
		}
		return strings.Join(parts, " ")
	}

	for _, pf := range ctx.PortForwards {
		parts = append(parts, formatPFEntry(pf, m.opts.NoColor))
	}
	return strings.Join(parts, " ")
}

func formatPFEntry(pf PortForwardLine, noColor bool) string {
	mark := "✓"
	color := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	if !pf.Connected {
		mark = "✗"
		color = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	}
	if noColor {
		if pf.Connected {
			mark = "+"
		} else {
			mark = "-"
		}
		return fmt.Sprintf("%s %s:%d", mark, pf.Name, pf.LocalPort)
	}
	return color.Render(mark) + fmt.Sprintf(" %s:%d", pf.Name, pf.LocalPort)
}
