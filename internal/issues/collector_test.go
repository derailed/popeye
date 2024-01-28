// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package issues

import (
	"context"
	"errors"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
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
				New(types.NewGVR("blee"), Root, rules.InfoLevel, "blee"),
				New(types.NewGVR("blee"), Root, rules.WarnLevel, "blee"),
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := NewCollector(nil, makeConfig(t))
			c.addIssue("fred", u.issues...)

			assert.Equal(t, u.e, c.NoConcerns("fred"))
		})
	}
}

func TestMaxSeverity(t *testing.T) {
	uu := map[string]struct {
		issues   []Issue
		section  string
		severity rules.Level
		count    int
	}{
		"noIssue": {
			section:  Root,
			severity: rules.OkLevel,
			count:    0,
		},
		"mix": {
			issues: []Issue{
				New(types.NewGVR("fred"), Root, rules.InfoLevel, "blee"),
				New(types.NewGVR("fred"), Root, rules.WarnLevel, "blee"),
			},
			section:  Root,
			severity: rules.WarnLevel,
			count:    2,
		},
		"same": {
			issues: []Issue{
				New(types.NewGVR("fred"), Root, rules.InfoLevel, "blee"),
				New(types.NewGVR("fred"), Root, rules.InfoLevel, "blee"),
			},
			section:  Root,
			severity: rules.InfoLevel,
			count:    2,
		},
		"error": {
			issues: []Issue{
				New(types.NewGVR("fred"), Root, rules.ErrorLevel, "blee"),
				New(types.NewGVR("fred"), Root, rules.InfoLevel, "blee"),
				New(types.NewGVR("fred"), Root, rules.InfoLevel, "blee"),
			},
			section:  Root,
			severity: rules.ErrorLevel,
			count:    3,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := NewCollector(nil, makeConfig(t))
			c.addIssue(u.section, u.issues...)

			assert.Equal(t, u.count, len(c.outcomes[u.section]))
			assert.Equal(t, u.severity, c.MaxSeverity(u.section))
		})
	}
}

func TestAddErr(t *testing.T) {
	uu := map[string]struct {
		errors []error
		fqn    string
		count  int
	}{
		"one": {
			errors: []error{
				errors.New("blee"),
			},
			fqn:   Root,
			count: 1,
		},
		"many": {
			errors: []error{
				errors.New("blee"),
				errors.New("duh"),
			},
			fqn:   Root,
			count: 2,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := NewCollector(nil, makeConfig(t))
			ctx := makeContext("errors", u.fqn, "")
			c.AddErr(ctx, u.errors...)

			assert.Equal(t, u.count, len(c.outcomes[u.fqn]))
			assert.Equal(t, rules.ErrorLevel, c.MaxSeverity(u.fqn))
		})
	}
}

func TestAddCode(t *testing.T) {
	uu := map[string]struct {
		code  rules.ID
		fqn   string
		args  []interface{}
		level rules.Level
		e     string
	}{
		"No params": {
			code:  100,
			fqn:   Root,
			level: rules.ErrorLevel,
			e:     "[POP-100] Untagged docker image in use",
		},
		"Params": {
			code:  108,
			fqn:   Root,
			level: rules.InfoLevel,
			args:  []interface{}{80},
			e:     "[POP-108] Unnamed port 80",
		},
		"Dud!": {
			code:  0,
			fqn:   Root,
			level: rules.InfoLevel,
			args:  []interface{}{80},
			e:     "[POP-108] Unnamed port 80",
		},
		"Issue 169": {
			code:  1102,
			fqn:   Root,
			level: rules.InfoLevel,
			args:  []interface{}{"123", "test-port"},
			e:     "[POP-1102] Use of target port #123 for service port test-port. Prefer named port",
		},
	}

	for k := range uu {
		u, key := uu[k], k
		t.Run(k, func(t *testing.T) {
			c := NewCollector(loadCodes(t), makeConfig(t))
			ctx := makeContext("test", u.fqn, "")
			if key == "Dud!" {
				subCode := func() {
					c.AddCode(ctx, u.code, u.args...)
				}
				assert.Panics(t, subCode, "blee")
			} else {
				c.AddCode(ctx, u.code, u.args...)
				assert.Equal(t, u.e, c.outcomes[u.fqn][0].Message)
				assert.Equal(t, u.level, c.outcomes[u.fqn][0].Level)
			}
		})
	}
}

func TestAddSubCode(t *testing.T) {
	uu := map[string]struct {
		code           rules.ID
		section, group string
		args           []interface{}
		level          rules.Level
		e              string
	}{
		"No params": {
			code:    100,
			section: Root,
			group:   "blee",
			level:   rules.ErrorLevel,
			e:       "[POP-100] Untagged docker image in use",
		},
		"Params": {
			code:    108,
			section: Root,
			group:   "blee",
			level:   rules.InfoLevel,
			args:    []interface{}{80},
			e:       "[POP-108] Unnamed port 80",
		},
		"Dud!": {
			code:    0,
			section: Root,
			group:   "blee",
			level:   rules.InfoLevel,
			args:    []interface{}{80},
			e:       "[POP-108] Unnamed port 80",
		},
	}

	for k := range uu {
		u, key := uu[k], k
		t.Run(k, func(t *testing.T) {
			c := NewCollector(loadCodes(t), makeConfig(t))
			c.InitOutcome(u.section)
			ctx := makeContext("test", u.section, u.group)

			if key == "Dud!" {
				subCode := func() {
					c.AddSubCode(ctx, u.code, u.args)
				}
				assert.Panics(t, subCode, "blee")
			} else {
				c.AddSubCode(ctx, u.code, u.args...)

				assert.Equal(t, u.e, c.Outcome()[u.section][0].Message)
				assert.Equal(t, u.level, c.Outcome()[u.section][0].Level)
			}
		})
	}
}

// Helpers...

func makeContext(section, fqn, group string) context.Context {
	return context.WithValue(context.Background(), internal.KeyRunInfo, internal.RunInfo{
		Section: section,
		Group:   group,
		Spec:    rules.Spec{FQN: fqn},
	})
}

func makeConfig(t *testing.T) *config.Config {
	c, err := config.NewConfig(config.NewFlags())
	assert.Nil(t, err)
	return c
}

func loadCodes(t *testing.T) *Codes {
	codes, err := LoadCodes()
	assert.Nil(t, err)

	return codes
}
