package output

import (
	"strings"
)

const (
	ansiBoldCyan  = "\033[1;36m"
	ansiBoldWhite = "\033[1;37m"
	ansiDim       = "\033[2m"
	ansiResetBold = "\033[0m"

	sectionSeparatorStr = "────────────────────────────────────────"
)

// ColorizeYAMLSections scans s line-by-line and applies ANSI coloring:
//   - Top-level section headers (^word:$) → bold cyan
//   - Indented key: value lines → bold-white key, dim separator, green value
//   - Indented key: <sentinel> lines → bold-white key, dim separator, dim value
//   - Indented key: (no value) lines → bold-white key, dim colon
//
// When noColor is true the input is returned unchanged.
func ColorizeYAMLSections(s string, noColor bool) string {
	if noColor {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if isSectionHeader(line) {
			lines[i] = ansiBoldCyan + line + ansiResetBold
			continue
		}

		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == line || trimmed == "" {
			continue
		}
		indent := line[:len(line)-len(trimmed)]

		colonIdx := strings.IndexByte(trimmed, ':')
		if colonIdx < 0 {
			continue
		}

		key := trimmed[:colonIdx]
		rest := trimmed[colonIdx+1:]

		if strings.TrimSpace(rest) == "" {
			lines[i] = indent + ansiBoldWhite + key + ansiResetBold + ansiDim + ":" + ansiResetBold
			continue
		}

		// Preserve the space after ":" when present.
		var sep, value string
		if strings.HasPrefix(rest, " ") {
			sep = ": "
			value = rest[1:]
		} else {
			sep = ":"
			value = rest
		}

		var coloredValue string
		if strings.HasPrefix(value, "<") && strings.HasSuffix(value, ">") {
			coloredValue = ansiDim + value + ansiResetBold
		} else {
			coloredValue = ansiGreen + value + ansiResetBold
		}

		lines[i] = indent + ansiBoldWhite + key + ansiResetBold + ansiDim + sep + ansiResetBold + coloredValue
	}
	return strings.Join(lines, "\n")
}

// isSectionHeader reports whether line matches the pattern ^(\w[\w-]*):\s*$
// — a top-level YAML key with no inline value.
func isSectionHeader(line string) bool {
	if len(line) == 0 {
		return false
	}
	if !isWordChar(rune(line[0])) {
		return false
	}
	colonIdx := strings.Index(line, ":")
	if colonIdx < 0 {
		return false
	}
	for _, r := range line[:colonIdx] {
		if !isWordChar(r) && r != '-' {
			return false
		}
	}
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
