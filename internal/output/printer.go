// Package output provides multi-format output rendering for the HyperFleet CLI.
// Printer dispatches --output json|table|yaml and handles colored dots,
// dynamic column ordering, and colored JSON output.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// Printer renders output in json, table, or yaml format.
type Printer struct {
	format  string
	noColor bool
	w       io.Writer
	errW    io.Writer
}

// NewPrinter creates a Printer.
// format: "json", "table", or "yaml". noColor disables ANSI codes.
// If w is nil, os.Stdout is used; errW defaults to os.Stderr.
func NewPrinter(format string, noColor bool, w io.Writer, errW io.Writer) *Printer {
	if w == nil {
		w = os.Stdout
	}
	if errW == nil {
		errW = os.Stderr
	}
	if !noColor {
		// Auto-disable color for non-TTY stdout
		if f, ok := w.(*os.File); ok {
			if !term.IsTerminal(int(f.Fd())) {
				noColor = true
			}
		} else {
			// Non-file writer (e.g., bytes.Buffer in tests) → no color
			noColor = true
		}
		// Respect NO_COLOR env var
		if os.Getenv("NO_COLOR") != "" {
			noColor = true
		}
	}
	return &Printer{format: format, noColor: noColor, w: w, errW: errW}
}

// Print renders v as JSON or YAML depending on the format.
func (p *Printer) Print(v any) error {
	switch p.format {
	case "yaml":
		return p.printYAML(v)
	default:
		return p.printJSON(v)
	}
}

func (p *Printer) printJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if !p.noColor {
		b = colorizeJSON(b)
	}
	_, err = fmt.Fprintf(p.w, "%s\n", b)
	return err
}

func (p *Printer) printYAML(v any) error {
	enc := yaml.NewEncoder(p.w)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return err
	}
	return enc.Close()
}

// PrintTable renders headers and rows as an aligned table.
// Headers are uppercased. Headers longer than 10 characters are wrapped
// across multiple stacked lines so the table stays narrow.
// Column widths are computed from visible display widths (ANSI codes excluded)
// so colored data cells remain visually aligned under their headers.
func (p *Printer) PrintTable(headers []string, rows [][]string) error {
	const maxHeaderWidth = 10
	const colSep = 2

	// Uppercase and wrap each header.
	wrapped := make([][]string, len(headers))
	maxHeaderLines := 0
	for i, h := range headers {
		lines := WrapHeader(strings.ToUpper(h), maxHeaderWidth)
		wrapped[i] = lines
		if len(lines) > maxHeaderLines {
			maxHeaderLines = len(lines)
		}
	}

	numCols := len(headers)

	// Build the full row list: header rows first, then data rows.
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

	// Compute column widths using display widths (ANSI escape codes do not
	// contribute to visible width, so strip them before measuring).
	colWidths := make([]int, numCols)
	for _, row := range allRows {
		for col, cell := range row {
			if w := displayWidth(cell); w > colWidths[col] {
				colWidths[col] = w
			}
		}
	}

	// Render: write each cell followed by padding to reach column width + separator.
	// The last column in each row gets no trailing padding.
	for _, row := range allRows {
		for col, cell := range row {
			fmt.Fprint(p.w, cell)
			if col < numCols-1 {
				pad := colWidths[col] - displayWidth(cell) + colSep
				fmt.Fprint(p.w, strings.Repeat(" ", pad))
			}
		}
		fmt.Fprintln(p.w)
	}
	return nil
}

// stripANSI removes ANSI escape sequences from s using a simple state machine.
func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		switch {
		case r == '\033':
			inEsc = true
		case inEsc:
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// runeWidth returns the number of terminal columns a rune occupies.
// Most runes are 1 column wide; CJK ideographs, full-width forms, and many
// emoji are 2 columns wide. This avoids a third-party dependency by covering
// the Unicode ranges that are actually wide (East Asian Width = Wide or
// Fullwidth) plus the common Emoji block used in this CLI (e.g. U+274C ❌).
func runeWidth(r rune) int {
	switch {
	case r >= 0x1100 && r <= 0x115F: // Hangul Jamo
		return 2
	case r >= 0x2E80 && r <= 0x303E: // CJK Radicals, Kangxi, etc.
		return 2
	case r >= 0x3041 && r <= 0x33BF: // Hiragana … CJK Compatibility
		return 2
	case r >= 0x33FF && r <= 0xA4CF: // CJK Unified Ideographs Extension A, etc.
		return 2
	case r >= 0xA960 && r <= 0xA97F: // Hangul Jamo Extended-A
		return 2
	case r >= 0xAC00 && r <= 0xD7FF: // Hangul Syllables
		return 2
	case r >= 0xF900 && r <= 0xFAFF: // CJK Compatibility Ideographs
		return 2
	case r >= 0xFE10 && r <= 0xFE19: // Vertical forms
		return 2
	case r >= 0xFE30 && r <= 0xFE6F: // CJK Compatibility Forms, Small Forms
		return 2
	case r >= 0xFF01 && r <= 0xFF60: // Fullwidth Latin, Katakana, etc.
		return 2
	case r >= 0xFFE0 && r <= 0xFFE6: // Fullwidth signs
		return 2
	case r >= 0x1B000 && r <= 0x1B0FF: // Kana Supplement
		return 2
	case r >= 0x1F004 && r <= 0x1F0CF: // Playing cards, Mahjong
		return 2
	case r >= 0x1F200 && r <= 0x1FFFF: // Enclosed Ideographic, Misc symbols, Emoji
		return 2
	case r >= 0x20000 && r <= 0x2FFFD: // CJK Unified Ideographs Extension B–F
		return 2
	case r >= 0x30000 && r <= 0x3FFFD: // CJK Unified Ideographs Extension G+
		return 2
	case r >= 0x2600 && r <= 0x27BF: // Misc Symbols, Dingbats (includes ❌ U+274C)
		return 2
	default:
		return 1
	}
}

// displayWidth returns the number of terminal columns s occupies (ANSI codes excluded).
// It accounts for wide Unicode characters (emoji, CJK) that occupy 2 columns each.
func displayWidth(s string) int {
	w := 0
	for _, r := range stripANSI(s) {
		w += runeWidth(r)
	}
	return w
}

// WrapHeader splits a header string into lines of at most maxWidth characters.
// Splitting prefers underscore word boundaries; single tokens longer than
// maxWidth are hard-broken at maxWidth.
func WrapHeader(s string, maxWidth int) []string {
	if len(s) <= maxWidth {
		return []string{s}
	}

	tokens := strings.Split(s, "_")

	var lines []string
	current := ""
	for _, tok := range tokens {
		// Hard-break any single token that exceeds maxWidth.
		for len(tok) > maxWidth {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			lines = append(lines, tok[:maxWidth])
			tok = tok[maxWidth:]
		}
		if tok == "" {
			continue
		}

		var candidate string
		if current == "" {
			candidate = tok
		} else {
			candidate = current + "_" + tok
		}

		if len(candidate) <= maxWidth {
			current = candidate
		} else {
			lines = append(lines, current)
			current = tok
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// Warn writes a [WARN] message to stderr.
func (p *Printer) Warn(msg string) {
	fmt.Fprintf(p.errW, "[WARN] %s\n", msg)
}

// Info writes an [INFO] message to stderr.
func (p *Printer) Info(msg string) {
	fmt.Fprintf(p.errW, "[INFO] %s\n", msg)
}

// Error writes an [ERROR] message to stderr.
func (p *Printer) Error(msg string) {
	fmt.Fprintf(p.errW, "[ERROR] %s\n", msg)
}

// colorizeJSON applies ANSI color codes to JSON bytes.
// Keys → cyan, string values → green, numbers → yellow, true → green, false → red, null → dim.
func colorizeJSON(src []byte) []byte {
	const (
		cyan    = "\033[36m"
		green   = "\033[32m"
		red     = "\033[31m"
		yellow  = "\033[33m"
		dim     = "\033[2m"
		reset   = "\033[0m"
	)

	var out bytes.Buffer
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.UseNumber()

	// We need to re-encode the pre-indented JSON with color codes.
	// Strategy: re-marshal from the indented bytes using a token-by-token walk.
	// Since the input is already indented JSON, we do a simple string-level
	// colorization by scanning the bytes.
	out = colorizeJSONBytes(src, cyan, green, red, yellow, dim, reset)
	return out.Bytes()
}

// colorizeJSONBytes performs a simple byte-level scan to apply colors to JSON tokens.
func colorizeJSONBytes(src []byte, keyCyan, strGreen, falseRed, numYellow, nullDim, reset string) bytes.Buffer {
	var out bytes.Buffer
	i := 0
	n := len(src)
	inKey := false // true when we're looking for a key (after { or ,)
	_ = inKey

	// Track structure depth to know if next string is a key or value.
	// After '{' or ',' at object level, the next string is a key.
	type frame struct{ isObj bool }
	stack := []frame{}

	expectKey := false

	for i < n {
		ch := src[i]

		// Skip whitespace, emit as-is
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			out.WriteByte(ch)
			i++
			continue
		}

		switch ch {
		case '{':
			stack = append(stack, frame{isObj: true})
			expectKey = true
			out.WriteByte(ch)
			i++
		case '[':
			stack = append(stack, frame{isObj: false})
			expectKey = false
			out.WriteByte(ch)
			i++
		case '}':
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			expectKey = false
			out.WriteByte(ch)
			i++
		case ']':
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			out.WriteByte(ch)
			i++
		case ':':
			expectKey = false
			out.WriteByte(ch)
			i++
		case ',':
			if len(stack) > 0 && stack[len(stack)-1].isObj {
				expectKey = true
			}
			out.WriteByte(ch)
			i++
		case '"':
			// Find end of string
			j := i + 1
			for j < n {
				if src[j] == '\\' {
					j += 2
					continue
				}
				if src[j] == '"' {
					j++
					break
				}
				j++
			}
			token := src[i:j]
			if expectKey {
				out.WriteString(keyCyan)
				out.Write(token)
				out.WriteString(reset)
			} else {
				out.WriteString(strGreen)
				out.Write(token)
				out.WriteString(reset)
			}
			i = j
		default:
			// Number, bool, null — collect until delimiter
			j := i
			for j < n && src[j] != ',' && src[j] != '}' && src[j] != ']' && src[j] != '\n' && src[j] != ' ' {
				j++
			}
			token := strings.TrimSpace(string(src[i:j]))
			switch {
			case token == "true":
				out.WriteString(strGreen + token + reset)
			case token == "false":
				out.WriteString(falseRed + token + reset)
			case token == "null":
				out.WriteString(nullDim + token + reset)
			default:
				// number
				out.WriteString(numYellow + token + reset)
			}
			i = j
		}
	}
	return out
}
