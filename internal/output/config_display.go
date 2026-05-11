package output

import (
	"strings"
)

const (
	ansiBoldCyan  = "\033[1;36m"
	ansiResetBold = "\033[0m"

	sectionSeparatorStr = "────────────────────────────────────────"
)

// ColorizeYAMLSections scans s line-by-line and wraps top-level YAML section-header
// lines (matching ^(\w[\w-]*):\s*$) with bold-cyan ANSI codes.
// When noColor is true the input is returned unchanged.
func ColorizeYAMLSections(s string, noColor bool) string {
	if noColor {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if isSectionHeader(line) {
			lines[i] = ansiBoldCyan + line + ansiResetBold
		}
	}
	return strings.Join(lines, "\n")
}

// isSectionHeader reports whether line matches the pattern ^(\w[\w-]*):\s*$
// — a top-level YAML key with no inline value.
func isSectionHeader(line string) bool {
	if len(line) == 0 {
		return false
	}
	// Must start with a word character.
	if !isWordChar(rune(line[0])) {
		return false
	}
	// Find the colon.
	colonIdx := strings.Index(line, ":")
	if colonIdx < 0 {
		return false
	}
	// All characters before the colon must be word chars or '-'.
	for _, r := range line[:colonIdx] {
		if !isWordChar(r) && r != '-' {
			return false
		}
	}
	// Everything after the colon must be whitespace only.
	rest := line[colonIdx+1:]
	return strings.TrimRight(rest, " \t") == ""
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// SectionSeparator returns a 40-character box-drawing separator line.
// The line carries no ANSI codes in either mode; noColor is accepted for
// API symmetry with other output helpers.
func SectionSeparator(noColor bool) string {
	_ = noColor
	return sectionSeparatorStr
}
