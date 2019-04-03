package linter

import (
	"fmt"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/rs/zerolog"
)

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

// ----------------------------------------------------------------------------

type (
	// Issue indicates a potential linter issue.
	Issue interface {
		Severity() Level
		Description() string
	}

	// Issues a collection of linter issues.
	Issues map[string][]Issue

	// Linter describes a lint resource.
	Linter struct {
		client *k8s.Client
		log    *zerolog.Logger
		issues Issues
	}
)

// NewLinter returns a new linter.
func newLinter(c *k8s.Client, l *zerolog.Logger) *Linter {
	return &Linter{client: c, log: l, issues: Issues{}}
}

// MaxSeverity scans the lint messages and return the highest severity.
func (l *Linter) MaxSeverity(res string) Level {
	max := OkLevel
	for _, issue := range l.issues[res] {
		if issue.Severity() > max {
			max = issue.Severity()
		}
	}
	return max
}

// NoIssues return true if not lint errors were detected. False otherwize
func (l *Linter) NoIssues(res string) bool {
	return len(l.issues[res]) == 0
}

// Issues returns a collection of linter issues.
func (l *Linter) Issues() Issues {
	return l.issues
}

func (l *Linter) initIssues(res string) {
	var issues []Issue
	l.issues[res] = issues
}

func (l *Linter) addIssuesMap(res string, issues Issues) {
	for k, v := range issues {
		for _, i := range v {
			err := Error{
				severity:    i.Severity(),
				description: fmt.Sprintf("%s%s%s", k, Delimiter, i.Description()),
			}
			l.issues[res] = append(l.issues[res], err)
		}
	}
}

func (l *Linter) addErrors(res string, errs []error) {
	for _, e := range errs {
		l.addIssue(res, ErrorLevel, e.Error())
	}
}

func (l *Linter) addError(res string, err error) {
	l.addIssue(res, ErrorLevel, err.Error())
}

func (l *Linter) addIssue(res string, level Level, msg string) {
	l.addIssues(res, NewError(level, msg))
}

func (l *Linter) addIssuef(res string, level Level, format string, args ...interface{}) {
	l.addIssues(res, NewErrorf(level, format, args...))
}

func (l *Linter) addIssues(res string, issues ...Issue) {
	l.issues[res] = append(l.issues[res], issues...)
}
