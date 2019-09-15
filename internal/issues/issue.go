package issues

import "fmt"

const (
	// OkLevel denotes no sanitizing issues.
	OkLevel Level = iota
	// InfoLevel denotes a FYI issues.
	InfoLevel
	// WarnLevel denotes a warning issue.
	WarnLevel
	// ErrorLevel denotes a serious issue.
	ErrorLevel
)

// Blank issue
var Blank = Issue{}

type (
	// Level tracks sanitizer message level.
	Level int

	// Issue tracks a sanitizer issui.
	Issue struct {
		Group   string `yaml:"group" json:"group"`
		Level   Level  `yaml:"level" json:"level"`
		Message string `yaml:"message" json:"message"`
	}
)

// New returns a new lint issue.
func New(group string, level Level, description string) Issue {
	return Issue{Group: group, Level: level, Message: description}
}

// Newf returns a new lint issue using a formatter.
func Newf(group string, level Level, format string, args ...interface{}) Issue {
	return New(group, level, fmt.Sprintf(format, args...))
}

// Blank checks if an issue is blank.
func (i Issue) Blank() bool {
	return i == Blank
}

// IsSubIssue checks if error is a sub error.
func (i Issue) IsSubIssue() bool {
	return i.Group != Root
}

// LevelToStr returns a severity level as a string.
func LevelToStr(l Level) string {
	switch l {
	case ErrorLevel:
		return "error"
	case WarnLevel:
		return "warn"
	case InfoLevel:
		return "info"
	default:
		return "ok"
	}
}
