package issues

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoConcerns(t *testing.T) {
	uu := map[string]struct {
		issues []Issue
		e      bool
	}{
		"noIssue": {
			e: true,
		},
		"issues": {
			issues: []Issue{
				New(Root, InfoLevel, "blee"),
				New(Root, WarnLevel, "blee"),
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			c := NewCollector(nil)
			c.addIssue("fred", u.issues...)

			assert.Equal(t, u.e, c.NoConcerns("fred"))
		})
	}
}

func TestMaxSeverity(t *testing.T) {
	uu := map[string]struct {
		issues   []Issue
		section  string
		severity Level
		count    int
	}{
		"noIssue": {
			section:  Root,
			severity: OkLevel,
			count:    0,
		},
		"mix": {
			issues: []Issue{
				New(Root, InfoLevel, "blee"),
				New(Root, WarnLevel, "blee"),
			},
			section:  Root,
			severity: WarnLevel,
			count:    2,
		},
		"same": {
			issues: []Issue{
				New(Root, InfoLevel, "blee"),
				New(Root, InfoLevel, "blee"),
			},
			section:  Root,
			severity: InfoLevel,
			count:    2,
		},
		"error": {
			issues: []Issue{
				New(Root, ErrorLevel, "blee"),
				New(Root, InfoLevel, "blee"),
				New(Root, InfoLevel, "blee"),
			},
			section:  Root,
			severity: ErrorLevel,
			count:    3,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			c := NewCollector(nil)
			c.addIssue(u.section, u.issues...)

			assert.Equal(t, u.count, len(c.outcomes[u.section]))
			assert.Equal(t, u.severity, c.MaxSeverity(u.section))
		})
	}
}

func TestAddErr(t *testing.T) {
	uu := map[string]struct {
		errors  []error
		section string
		count   int
	}{
		"one": {
			errors: []error{
				errors.New("blee"),
			},
			section: Root,
			count:   1,
		},
		"many": {
			errors: []error{
				errors.New("blee"),
				errors.New("duh"),
			},
			section: Root,
			count:   2,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			c := NewCollector(nil)
			c.AddErr(u.section, u.errors...)

			assert.Equal(t, u.count, len(c.outcomes[u.section]))
			assert.Equal(t, ErrorLevel, c.MaxSeverity(u.section))
		})
	}
}

func TestAddCode(t *testing.T) {
	uu := map[string]struct {
		code    ID
		section string
		args    []interface{}
		level   Level
		e       string
	}{
		"No params": {
			code:    100,
			section: Root,
			level:   ErrorLevel,
			e:       "[POP-100] Untagged docker image in use",
		},
		"Params": {
			code:    108,
			section: Root,
			level:   InfoLevel,
			args:    []interface{}{80},
			e:       "[POP-108] Unamed port 80",
		},
		"Dud!": {
			code:    0,
			section: Root,
			level:   InfoLevel,
			args:    []interface{}{80},
			e:       "[POP-108] Unamed port 80",
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			c := NewCollector(loadCodes(t))

			if k == "Dud!" {
				subCode := func() {
					c.AddCode(u.code, u.section, u.args...)
				}
				assert.Panics(t, subCode, "blee")
			} else {
				c.AddCode(u.code, u.section, u.args...)

				assert.Equal(t, u.e, c.outcomes[u.section][0].Message)
				assert.Equal(t, u.level, c.outcomes[u.section][0].Level)
			}
		})
	}
}

func TestAddSubCode(t *testing.T) {
	uu := map[string]struct {
		code           ID
		section, group string
		args           []interface{}
		level          Level
		e              string
	}{
		"No params": {
			code:    100,
			section: Root,
			group:   "blee",
			level:   ErrorLevel,
			e:       "[POP-100] Untagged docker image in use",
		},
		"Params": {
			code:    108,
			section: Root,
			group:   "blee",
			level:   InfoLevel,
			args:    []interface{}{80},
			e:       "[POP-108] Unamed port 80",
		},
		"Dud!": {
			code:    0,
			section: Root,
			group:   "blee",
			level:   InfoLevel,
			args:    []interface{}{80},
			e:       "[POP-108] Unamed port 80",
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			c := NewCollector(loadCodes(t))
			c.InitOutcome(u.section)

			if k == "Dud!" {
				subCode := func() {
					c.AddSubCode(u.code, u.section, u.group, u.args)
				}
				assert.Panics(t, subCode, "blee")
			} else {
				c.AddSubCode(u.code, u.section, u.group, u.args...)

				assert.Equal(t, u.e, c.Outcome()[u.section][0].Message)
				assert.Equal(t, u.level, c.Outcome()[u.section][0].Level)
			}
		})
	}
}

func loadCodes(t *testing.T) *Codes {
	codes, err := LoadCodes("../../assets/codes.yml")
	assert.Nil(t, err)
	return codes
}
