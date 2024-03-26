// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package issues

import (
	"context"
	"fmt"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog/log"
)

const errCode = 666

// Collector tracks linter issues and codes.
type Collector struct {
	*config.Config

	outcomes Outcome
	codes    *Codes
}

// NewCollector returns a new issue collector.
func NewCollector(codes *Codes, cfg *config.Config) *Collector {
	return &Collector{Config: cfg, outcomes: Outcome{}, codes: codes}
}

// Outcome returns scan outcome.
func (c *Collector) Outcome() Outcome {
	return c.outcomes
}

// InitOutcome creates a places holder for potential issues.
func (c *Collector) InitOutcome(fqn string) {
	c.outcomes[fqn] = Issues{}
}

func (c *Collector) CloseOutcome(ctx context.Context, fqn string, cos []string) {
	if c.NoConcerns(fqn) && c.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn, cos) {
		c.ClearOutcome(fqn)
	}
}

// ClearOutcome delete all fqn related issues.
func (c *Collector) ClearOutcome(fqn string) {
	delete(c.outcomes, fqn)
}

// NoConcerns returns true if scan is successful.
func (c *Collector) NoConcerns(fqn string) bool {
	return len(c.outcomes[fqn]) == 0
}

// MaxSeverity return the highest severity level for the given section.
func (c *Collector) MaxSeverity(fqn string) rules.Level {
	return c.outcomes.MaxSeverity(fqn)
}

// AddSubCode add a sub error code.
func (c *Collector) AddSubCode(ctx context.Context, code rules.ID, args ...interface{}) {
	run := internal.MustExtractRunInfo(ctx)
	co, ok := c.codes.Glossary[code]
	if !ok {
		log.Error().Err(fmt.Errorf("No code with ID %d", code)).Msg("AddSubCode failed")
	}
	if co.Severity < rules.Level(c.Config.LintLevel) {
		return
	}

	run.Spec.GVR, run.Spec.Code = run.SectionGVR, code
	if !c.Match(run.Spec) {
		c.addIssue(run.Spec.FQN, New(run.GroupGVR, run.Group, co.Severity, co.Format(code, args...)))
	}
}

// AddCode add an error code.
func (c *Collector) AddCode(ctx context.Context, code rules.ID, args ...interface{}) {
	run := internal.MustExtractRunInfo(ctx)
	co, ok := c.codes.Glossary[code]
	if !ok {
		// BOZO!! refact once codes are in!!
		panic(fmt.Errorf("no codes found with id %d", code))
	}
	if co.Severity < rules.Level(c.Config.LintLevel) {
		return
	}

	run.Spec.GVR, run.Spec.Code = run.SectionGVR, code
	if !c.Match(run.Spec) {
		c.addIssue(run.Spec.FQN, New(run.SectionGVR, Root, co.Severity, co.Format(code, args...)))
	}
}

// AddErr adds a collection of errors.
func (c *Collector) AddErr(ctx context.Context, errs ...error) {
	run := internal.MustExtractRunInfo(ctx)
	if c.codes == nil {
		for _, e := range errs {
			c.addIssue(run.Spec.FQN, New(run.SectionGVR, Root, rules.ErrorLevel, e.Error()))
		}
		return
	}

	co, ok := c.codes.Glossary[errCode]
	if !ok {
		// BOZO!! refact once codes are in!!
		panic(fmt.Errorf("no codes found with id %d", errCode))
	}
	for _, e := range errs {
		c.addIssue(run.Spec.FQN, New(run.SectionGVR, Root, rules.ErrorLevel, co.Format(errCode, e.Error())))
	}
}

// AddIssue adds 1 or more concerns to the collector.
func (c *Collector) addIssue(fqn string, concerns ...Issue) {
	if len(concerns) == 0 {
		return
	}
	c.outcomes[fqn] = append(c.outcomes[fqn], concerns...)
}
