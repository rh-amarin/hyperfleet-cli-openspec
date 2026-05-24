package output

import "testing"

func TestFormatTableMatchesPrintTableAlignment(t *testing.T) {
	green := "\033[32m●\033[0m"
	headers := []string{"id", "name", "status"}
	rows := [][]string{
		{"1", "cluster-a", green + " 1"},
		{"22", "cluster-b", green + " 2"},
	}

	ft := FormatTable(headers, rows)
	if len(ft.DataLines) != 2 {
		t.Fatalf("data lines = %d", len(ft.DataLines))
	}
	// NAME values should align vertically.
	a := indexVisible(ft.DataLines[0], "cluster-a")
	b := indexVisible(ft.DataLines[1], "cluster-b")
	if a != b {
		t.Errorf("columns misaligned: %q vs %q", ft.DataLines[0], ft.DataLines[1])
	}
}

func indexVisible(line, sub string) int {
	return len(stripANSI(line[:maxIndex(line, sub)+1]))
}

func maxIndex(line, sub string) int {
	i := 0
	for idx := 0; idx <= len(line)-len(sub); idx++ {
		if line[idx:idx+len(sub)] == sub {
			i = idx
		}
	}
	return i
}
