package output

import (
	"strings"
	"testing"
)

func TestColorizeYAMLSections_noColor(t *testing.T) {
	input := "hyperfleet:\n  api-url: http://localhost:8000\n  token: <not set>\n  context:\ndatabase:\n  host: localhost\n"
	got := ColorizeYAMLSections(input, true)
	if got != input {
		t.Errorf("noColor=true: expected output to equal input\ngot:  %q\nwant: %q", got, input)
	}
}

func TestColorizeYAMLSections_withColor_sectionHeaders(t *testing.T) {
	input := "hyperfleet:\n  api-url: http://localhost:8000\ndatabase:\n  host: localhost\n"
	got := ColorizeYAMLSections(input, false)

	for _, line := range strings.Split(got, "\n") {
		if line == "" {
			continue
		}
		plain := stripANSI(line)
		if isSectionHeader(plain) && !strings.Contains(line, ansiBoldCyan) {
			t.Errorf("section-header line missing bold-cyan escape: %q", line)
		}
	}
}

func TestColorizeYAMLSections_indentedKeyValue(t *testing.T) {
	input := "  api-url: http://example.com"
	got := ColorizeYAMLSections(input, false)
	if !strings.Contains(got, ansiBoldWhite) {
		t.Errorf("key segment missing bold-white: %q", got)
	}
	if !strings.Contains(got, ansiGreen) {
		t.Errorf("value segment missing green: %q", got)
	}
}

func TestColorizeYAMLSections_indentedSentinel(t *testing.T) {
	input := "  token: <not set>"
	got := ColorizeYAMLSections(input, false)
	// value must be dim, not green
	afterColon := got[strings.Index(got, ": ")+2:]
	if strings.Contains(afterColon, ansiGreen) {
		t.Errorf("sentinel value must not use green: %q", got)
	}
	if !strings.Contains(afterColon, ansiDim) {
		t.Errorf("sentinel value must use dim: %q", got)
	}
}

func TestColorizeYAMLSections_indentedKeyOnly(t *testing.T) {
	input := "  context:"
	got := ColorizeYAMLSections(input, false)
	if !strings.Contains(got, ansiBoldWhite) {
		t.Errorf("key segment missing bold-white: %q", got)
	}
	if strings.Contains(got, ansiGreen) {
		t.Errorf("key-only line must not contain green: %q", got)
	}
}

func TestColorizeYAMLSections_noColor_allTypes(t *testing.T) {
	lines := []string{
		"hyperfleet:",
		"  api-url: http://example.com",
		"  token: <not set>",
		"  context:",
	}
	for _, line := range lines {
		got := ColorizeYAMLSections(line, true)
		if got != line {
			t.Errorf("noColor=true: line changed\ngot:  %q\nwant: %q", got, line)
		}
		if strings.Contains(got, "\033[") {
			t.Errorf("noColor=true: ANSI code present in %q", got)
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
