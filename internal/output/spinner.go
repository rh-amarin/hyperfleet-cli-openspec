package output

import "time"

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// SpinnerFrame returns a braille spinner character at position tick%10.
func SpinnerFrame(tick int) string {
	return spinnerFrames[tick%10]
}

// IsActive returns true when lastReportTime is within 2×frequencySecs seconds of now.
// Returns false if lastReportTime is empty or not a valid RFC 3339 timestamp.
func IsActive(lastReportTime string, frequencySecs int) bool {
	if lastReportTime == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339, lastReportTime)
	if err != nil {
		return false
	}
	return time.Since(t) < time.Duration(2*frequencySecs)*time.Second
}
