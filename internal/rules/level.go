// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

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

// ToIssueLevel convert a string to a issue level.
func ToIssueLevel(level *string) Level {
	if level == nil || *level == "" {
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

func (l Level) ToHumanLevel() string {
	switch l {
	case OkLevel:
		return "ok"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	default:
		return "n/a"
	}
}
