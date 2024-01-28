// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package issues

import (
	"fmt"
	"regexp"

	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/types"
)

var codeRX = regexp.MustCompile(`\A\[POP-(\d+)\]`)

// Blank issue
var Blank = Issue{}

// Issue tracks a linter issue.
type Issue struct {
	Group   string      `yaml:"group" json:"group"`
	GVR     string      `yaml:"gvr" json:"gvr"`
	Level   rules.Level `yaml:"level" json:"level"`
	Message string      `yaml:"message" json:"message"`
}

// New returns a new lint issue.
func New(gvr types.GVR, group string, level rules.Level, description string) Issue {
	return Issue{GVR: gvr.String(), Group: group, Level: level, Message: description}
}

// Newf returns a new lint issue using a formatter.
func Newf(gvr types.GVR, group string, level rules.Level, format string, args ...interface{}) Issue {
	return New(gvr, group, level, fmt.Sprintf(format, args...))
}

func (i Issue) Code() (string, bool) {
	mm := codeRX.FindStringSubmatch(i.Message)
	if len(mm) < 2 {
		return "", false
	}

	return mm[1], true
}

// Dump for debugging.
func (i Issue) Dump() {
	fmt.Printf("  %s (%d) %s\n", i.GVR, i.Level, i.Message)
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
func LevelToStr(l rules.Level) string {
	// nolint:exhaustive
	switch l {
	case rules.ErrorLevel:
		return "error"
	case rules.WarnLevel:
		return "warn"
	case rules.InfoLevel:
		return "info"
	default:
		return "ok"
	}
}
