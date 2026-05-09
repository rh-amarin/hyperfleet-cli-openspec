package output

const (
	dot = "●"

	ansiGreen  = "\033[32m"
	ansiRed    = "\033[31m"
	ansiYellow = "\033[33m"
	ansiReset  = "\033[0m"
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
		return ansiGreen + dot + ansiReset
	case "False":
		if noColor {
			return "False"
		}
		return ansiRed + dot + ansiReset
	case "Unknown":
		if noColor {
			return "Unknown"
		}
		return ansiYellow + dot + ansiReset
	default:
		return "-"
	}
}
