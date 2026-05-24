package output

import "strings"

const (
	tableMaxHeaderWidth = 10
	tableColSep           = 2
)

// FormattedTable holds aligned header and data lines from FormatTable.
type FormattedTable struct {
	HeaderLines []string
	DataLines   []string
}

// FormatTable renders headers and rows as aligned lines using the same rules as PrintTable.
func FormatTable(headers []string, rows [][]string) FormattedTable {
	wrapped, maxHeaderLines, numCols := wrapTableHeaders(headers)

	allRows := make([][]string, 0, maxHeaderLines+len(rows))
	for line := 0; line < maxHeaderLines; line++ {
		row := make([]string, numCols)
		for col, lines := range wrapped {
			if line < len(lines) {
				row[col] = lines[line]
			}
		}
		allRows = append(allRows, row)
	}
	allRows = append(allRows, rows...)

	colWidths := tableColWidths(allRows)

	var ft FormattedTable
	for _, row := range allRows[:maxHeaderLines] {
		ft.HeaderLines = append(ft.HeaderLines, formatTableRow(row, colWidths))
	}
	for _, row := range allRows[maxHeaderLines:] {
		ft.DataLines = append(ft.DataLines, formatTableRow(row, colWidths))
	}
	return ft
}

// StripANSI removes ANSI escape sequences from s.
func StripANSI(s string) string {
	return stripANSI(s)
}

// DisplayWidth returns visible terminal columns for s (ANSI excluded, wide runes counted).
func DisplayWidth(s string) int {
	return displayWidth(s)
}

// MaxLineWidth returns the widest visible line width in lines.
func MaxLineWidth(lines []string) int {
	max := 0
	for _, line := range lines {
		if w := DisplayWidth(line); w > max {
			max = w
		}
	}
	return max
}

func wrapTableHeaders(headers []string) (wrapped [][]string, maxHeaderLines, numCols int) {
	numCols = len(headers)
	wrapped = make([][]string, numCols)
	for i, h := range headers {
		lines := WrapHeader(strings.ToUpper(h), tableMaxHeaderWidth)
		wrapped[i] = lines
		if len(lines) > maxHeaderLines {
			maxHeaderLines = len(lines)
		}
	}
	return wrapped, maxHeaderLines, numCols
}

func tableColWidths(allRows [][]string) []int {
	if len(allRows) == 0 {
		return nil
	}
	colWidths := make([]int, len(allRows[0]))
	for _, row := range allRows {
		for col, cell := range row {
			if w := displayWidth(cell); w > colWidths[col] {
				colWidths[col] = w
			}
		}
	}
	return colWidths
}

func formatTableRow(row []string, colWidths []int) string {
	var b strings.Builder
	for col, cell := range row {
		b.WriteString(cell)
		if col < len(row)-1 {
			pad := colWidths[col] - displayWidth(cell) + tableColSep
			if pad > 0 {
				b.WriteString(strings.Repeat(" ", pad))
			}
		}
	}
	return b.String()
}
