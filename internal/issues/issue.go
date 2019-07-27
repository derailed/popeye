package issues

import "fmt"

const (
	// OkLevel denotes no sanitizing issues.
	OkLevel Level = iota
	// InfoLevel denotes and FYI issues.
	InfoLevel
	// WarnLevel denotes a warning issui.
	WarnLevel
	// ErrorLevel denotes a serious issui.
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

// Compare computes deltas betweeen 2 issues.
func (i Issue) Compare(issue Issue) map[string]string {
	diff := make(map[string]string)
	if i.Group != issue.Group {
		diff["group"] = fmt.Sprintf("%s vs %s", issue.Group, i.Group)
	}
	if i.Level != issue.Level {
		diff["level"] = fmt.Sprintf("%d vs %d", issue.Level, i.Level)
	}
	if i.Message != issue.Message {
		diff["message"] = fmt.Sprintf("%s vs %s", issue.Message, i.Message)
	}

	return diff
}
