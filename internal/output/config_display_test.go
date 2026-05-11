package output

import (
	"strings"
	"testing"
)

func TestColorizeYAMLSections_noColor(t *testing.T) {
	input := "hyperfleet:\n  api-url: http://localhost:8000\ndatabase:\n  host: localhost\n"
	got := ColorizeYAMLSections(input, true)
	if got != input {
		t.Errorf("noColor=true: expected output to equal input\ngot:  %q\nwant: %q", got, input)
	}
}

func TestColorizeYAMLSections_withColor(t *testing.T) {
	input := "hyperfleet:\n  api-url: http://localhost:8000\ndatabase:\n  host: localhost\n"
	got := ColorizeYAMLSections(input, false)

	lines := strings.Split(got, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		isHeader := isSectionHeader(stripANSI(line))
		hasEscape := strings.Contains(line, "\033[")
		if isHeader && !hasEscape {
			t.Errorf("section-header line missing ANSI escape: %q", line)
		}
		if !isHeader && hasEscape {
			t.Errorf("value line must not contain ANSI escape: %q", line)
		}
	}
}

func TestSectionSeparator(t *testing.T) {
	for _, nc := range []bool{true, false} {
		s := SectionSeparator(nc)
		if s == "" {
			t.Errorf("noColor=%v: SectionSeparator returned empty string", nc)
		}
		if strings.Contains(s, "\033[") {
			t.Errorf("noColor=%v: SectionSeparator must not contain ANSI codes, got %q", nc, s)
		}
	}
}
