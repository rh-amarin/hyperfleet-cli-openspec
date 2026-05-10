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
	"text/tabwriter"

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

const maxHeaderWidth = 10

// PrintTable renders headers and rows as an aligned table.
// Headers are uppercased. Headers longer than maxHeaderWidth characters are
// wrapped across multiple rows so that the table does not become too wide.
func (p *Printer) PrintTable(headers []string, rows [][]string) error {
	tw := tabwriter.NewWriter(p.w, 0, 0, 2, ' ', 0)

	wrapped := make([][]string, len(headers))
	maxLines := 1
	for i, h := range headers {
		wrapped[i] = wrapHeader(strings.ToUpper(h), maxHeaderWidth)
		if len(wrapped[i]) > maxLines {
			maxLines = len(wrapped[i])
		}
	}

	for line := 0; line < maxLines; line++ {
		parts := make([]string, len(headers))
		for i, segments := range wrapped {
			if line < len(segments) {
				parts[i] = segments[line]
			}
		}
		fmt.Fprintln(tw, strings.Join(parts, "\t"))
	}

	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	return tw.Flush()
}

// wrapHeader splits s into lines of at most maxWidth characters, preferring
// to break at space or underscore boundaries. Tokens longer than maxWidth are
// split at the character level.
func wrapHeader(s string, maxWidth int) []string {
	if len(s) <= maxWidth {
		return []string{s}
	}
	tokens := strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == '_' })
	if len(tokens) == 0 {
		return chunkString(s, maxWidth)
	}
	var lines []string
	current := ""
	for _, tok := range tokens {
		if len(tok) > maxWidth {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			lines = append(lines, chunkString(tok, maxWidth)...)
			continue
		}
		switch {
		case current == "":
			current = tok
		case len(current)+1+len(tok) <= maxWidth:
			current += " " + tok
		default:
			lines = append(lines, current)
			current = tok
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// chunkString splits s into substrings of at most size characters.
func chunkString(s string, size int) []string {
	var chunks []string
	for len(s) > size {
		chunks = append(chunks, s[:size])
		s = s[size:]
	}
	if s != "" {
		chunks = append(chunks, s)
	}
	return chunks
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
