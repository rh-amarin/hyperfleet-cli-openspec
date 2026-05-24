package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
)

const (
	// Row highlight uses xterm color 33 to match the previous lipgloss selection style.
	ansiRowBG        = "\033[48;5;33m"
	ansiRowBGRestore = "\033[39;48;5;33m"
)

var headerStyle = lipgloss.NewStyle().Bold(true)

// formatMainTable renders aligned table lines, optionally clipping columns to maxWidth.
func formatMainTable(headers []string, rows [][]string, maxWidth int) output.FormattedTable {
	if maxWidth <= 0 || len(headers) == 0 {
		return output.FormatTable(headers, rows)
	}
	for n := len(headers); n >= 1; n-- {
		h, r := clipColumns(headers, rows, n)
		ft := output.FormatTable(h, r)
		all := append(ft.HeaderLines, ft.DataLines...)
		if output.MaxLineWidth(all) <= maxWidth {
			return ft
		}
	}
	return output.FormatTable(headers[:1], clipRows(rows, 1))
}

func clipColumns(headers []string, rows [][]string, n int) ([]string, [][]string) {
	h := headers[:n]
	return h, clipRows(rows, n)
}

func clipRows(rows [][]string, n int) [][]string {
	out := make([][]string, len(rows))
	for i, row := range rows {
		if len(row) >= n {
			out[i] = row[:n]
		} else {
			out[i] = row
		}
	}
	return out
}

func renderTableBlock(ft output.FormattedTable, selected, offsetY, visible, panelWidth int, noColor bool) string {
	var lines []string

	for _, line := range ft.HeaderLines {
		lines = append(lines, headerStyle.Render(line))
	}

	rowWidth := output.MaxLineWidth(ft.DataLines)
	if panelWidth > rowWidth {
		rowWidth = panelWidth
	}

	end := offsetY + visible
	if end > len(ft.DataLines) {
		end = len(ft.DataLines)
	}
	if offsetY > len(ft.DataLines) {
		offsetY = len(ft.DataLines)
	}

	for i := offsetY; i < end; i++ {
		line := ft.DataLines[i]
		if i == selected {
			line = highlightTableRow(line, rowWidth, noColor)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func highlightTableRow(line string, width int, noColor bool) string {
	if width <= 0 {
		width = output.DisplayWidth(line)
	}
	if noColor {
		plain := output.StripANSI(line)
		pad := width - output.DisplayWidth(plain)
		if pad > 0 {
			plain += strings.Repeat(" ", pad)
		}
		return "▸ " + plain
	}

	// Keep status-dot colors: each ANSIReset clears the row background, so re-apply it.
	highlighted := ansiRowBG + strings.ReplaceAll(line, output.ANSIReset, ansiRowBGRestore)
	pad := width - output.DisplayWidth(line)
	if pad > 0 {
		highlighted += strings.Repeat(" ", pad)
	}
	return highlighted + output.ANSIReset
}
