package linter

import "fmt"

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

type (
	// Level tracks lint check level.
	Level int

	// Error tracks a linter issue.
	Error struct {
		Level   Level  `yaml:"level"`
		Message string `yaml:"message"`
		Subs    Issues `yaml:"containers,omitempty"`
	}
)

// NewError returns a new lint issue.
func NewError(level Level, description string) *Error {
	return &Error{Level: level, Message: description, Subs: Issues{}}
}

// NewErrorf returns a new lint issue using a formatter.
func NewErrorf(level Level, format string, args ...interface{}) *Error {
	return NewError(level, fmt.Sprintf(format, args...))
}

// Severity returns the Level of the message.
func (e *Error) Severity() Level {
	return e.Level
}

// SetSeverity sets the severity level.
func (e *Error) SetSeverity(l Level) {
	e.Level = l
}

// Description returns the lint Message.
func (e *Error) Description() string {
	return e.Message
}

// HasSubIssues checks if error contains sub issues.
func (e *Error) HasSubIssues() bool {
	return len(e.Subs) > 0
}

// SubIssues returns the lint Message.
func (e *Error) SubIssues() Issues {
	return e.Subs
}
