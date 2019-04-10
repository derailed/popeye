package config

// Level tracks lint check level.
type Level int

const (
	// OkLevel denotes no linting issues.
	OkLevel Level = iota
	// InfoLevel denotes FIY linting issues.
	InfoLevel
	// WarnLevel denotes a warning issue.
	WarnLevel
	// ErrorLevel denotes a serious issue.
	ErrorLevel
)

// ToLintLevel convert a string to a level.
func ToLintLevel(level *string) Level {
	if !isSet(level) {
		return OkLevel
	}

	switch *level {
	case "ok":
		return OkLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return OkLevel
	}
}
