package linter

import "fmt"

const (
	// NoLevel describes an unclassified lint issue.
	NoLevel Level = iota
	// OkLevel denotes no linting issues.
	OkLevel
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

	// Error tracks a linter message.
	Error struct {
		severity    Level
		description string
	}

	// Issue indicates a potential linter issue.
	Issue interface {
		Severity() Level
		Description() string
	}

	// Issues a collection of linter issues.
	Issues []Issue
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

// ----------------------------------------------------------------------------

// Linter describes a linter resource.
type Linter struct {
	issues Issues
}

// MaxSeverity scans the lint messages and return the highest severity.
func (l *Linter) MaxSeverity() Level {
	max := OkLevel
	for _, i := range l.issues {
		if i.Severity() > max {
			max = i.Severity()
		}
	}
	return max
}

// NoIssues return true if not lint errors were detected. False otherwize
func (l *Linter) NoIssues() bool {
	return len(l.issues) == 0
}

// Issues returns a collection of linter issues.
func (l *Linter) Issues() Issues {
	return l.issues
}

func (l *Linter) addIssues(issues ...Issue) {
	l.issues = append(l.issues, issues...)
}

func (l *Linter) addIssue(level Level, msg string) {
	l.issues = append(l.issues, NewError(level, msg))
}

func (l *Linter) addIssuef(level Level, format string, args ...interface{}) {
	l.issues = append(l.issues, NewErrorf(level, format, args...))
}
