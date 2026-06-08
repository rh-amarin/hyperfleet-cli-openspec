package output

import "time"

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// DefaultWatchIntervalSecs is the assumed refresh interval for adapter activity checks
// when no watch interval is in use (e.g. one-shot hf rs).
const DefaultWatchIntervalSecs = 5

// SpinnerFrame returns a braille spinner character at position tick%10.
func SpinnerFrame(tick int) string {
	return spinnerFrames[tick%10]
}

// ActivityWindow returns 2× the effective refresh interval used for adapter activity checks.
// When frequencySecs is zero, DefaultWatchIntervalSecs is used.
func ActivityWindow(frequencySecs int) time.Duration {
	freq := frequencySecs
	if freq <= 0 {
		freq = DefaultWatchIntervalSecs
	}
	return time.Duration(2*freq) * time.Second
}

// IsActive returns true when lastReportTime is within the activity window of now.
// Returns false if lastReportTime is empty or not a valid RFC 3339 timestamp.
func IsActive(lastReportTime string, frequencySecs int) bool {
	if lastReportTime == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339, lastReportTime)
	if err != nil {
		return false
	}
	return time.Since(t) < ActivityWindow(frequencySecs)
}

// AdapterActivityPrefix returns the fixed 2-character activity slot prepended to every
// adapter table cell so column widths stay stable between watch and one-shot renders.
// Active adapters show the current spinner frame; inactive ones show two spaces.
func AdapterActivityPrefix(lastReportTime string, tick, frequencySecs int) string {
	if IsActive(lastReportTime, frequencySecs) {
		return SpinnerFrame(tick) + " "
	}
	return "  "
}
