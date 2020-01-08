package issues

import (
	"fmt"

	"github.com/derailed/popeye/pkg/config"
)

// Blank issue
var Blank = Issue{}

type (
	// Issue tracks a sanitizer issui.
	Issue struct {
		Group   string       `yaml:"group" json:"group"`
		Level   config.Level `yaml:"level" json:"level"`
		Message string       `yaml:"message" json:"message"`
	}
)

// New returns a new lint issue.
func New(group string, level config.Level, description string) Issue {
	return Issue{Group: group, Level: level, Message: description}
}

// Newf returns a new lint issue using a formatter.
func Newf(group string, level config.Level, format string, args ...interface{}) Issue {
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
func LevelToStr(l config.Level) string {
	switch l {
	case config.ErrorLevel:
		return "error"
	case config.WarnLevel:
		return "warn"
	case config.InfoLevel:
		return "info"
	default:
		return "ok"
	}
}
