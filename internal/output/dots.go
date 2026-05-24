package output

const (
	dot = "●"

	// ANSIReset is the standard SGR reset sequence used in status dots and spinners.
	ANSIReset = "\033[0m"

	ansiGreen  = "\033[32m"
	ansiRed    = "\033[31m"
	ansiYellow = "\033[33m"
)

// StatusDot returns a colored dot (or plain text) representing a condition status.
// status: "True" → green●, "False" → red●, "Unknown" → yellow●, "" → "-"
// noColor disables ANSI codes.
func StatusDot(status string, noColor bool) string {
	switch status {
	case "True":
		if noColor {
			return "True"
		}
		return ansiGreen + dot + ANSIReset
	case "False":
		if noColor {
			return "False"
		}
		return ansiRed + dot + ANSIReset
	case "Unknown":
		if noColor {
			return "Unknown"
		}
		return ansiYellow + dot + ANSIReset
	default:
		return "-"
	}
}
