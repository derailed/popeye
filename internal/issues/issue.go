package issues

import (
	"fmt"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/pkg/config"
)

// Blank issue
var Blank = Issue{}

type (
	// Issue tracks a sanitizer issui.
	Issue struct {
		Group   string       `yaml:"group" json:"group"`
		GVR     string       `yaml:"gvr" json:"gvr"`
		Level   config.Level `yaml:"level" json:"level"`
		Message string       `yaml:"message" json:"message"`
	}
)

// New returns a new lint issue.
func New(gvr client.GVR, group string, level config.Level, description string) Issue {
	return Issue{GVR: gvr.String(), Group: group, Level: level, Message: description}
}

// Newf returns a new lint issue using a formatter.
func Newf(gvr client.GVR, group string, level config.Level, format string, args ...interface{}) Issue {
	return New(gvr, group, level, fmt.Sprintf(format, args...))
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
	// nolint:exhaustive
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
