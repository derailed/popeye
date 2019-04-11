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

	// Delimiter indicates a sub section.
	Delimiter = "||"
)

type (
	// Level tracks lint check level.
	Level int

	// Error tracks a linter issue.
	Error struct {
		severity    Level
		description string
	}
)

// NewErrorf returns a new lint issue using a formatter.
func NewErrorf(level Level, format string, args ...interface{}) Error {
	return Error{severity: level, description: fmt.Sprintf(format, args...)}
}

// NewError returns a new lint issue.
func NewError(level Level, description string) Error {
	return Error{severity: level, description: description}
}

// Severity returns the severity of the message.
func (e Error) Severity() Level {
	return e.severity
}

// Description returns the lint description.
func (e Error) Description() string {
	return e.description
}
