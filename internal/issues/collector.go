package issues

import (
	"context"
	"fmt"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog/log"
)

// Collector represents a sanitizer issue container.
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

// ClearOutcome delete all fqn related issues.
func (c *Collector) ClearOutcome(fqn string) {
	delete(c.outcomes, fqn)
}

// NoConcerns returns true if scan is successful.
func (c *Collector) NoConcerns(fqn string) bool {
	return len(c.outcomes[fqn]) == 0
}

// MaxSeverity return the highest severity level foe the given section.
func (c *Collector) MaxSeverity(fqn string) config.Level {
	return c.outcomes.MaxSeverity(fqn)
}

// AddSubCode add a sub error code.
func (c *Collector) AddSubCode(ctx context.Context, code config.ID, args ...interface{}) {
	run := internal.MustExtractRunInfo(ctx)
	co, ok := c.codes.Glossary[code]
	if !ok {
		log.Error().Err(fmt.Errorf("No code with ID %d", code)).Msg("AddSubCode failed")
	}
	if !c.ShouldExclude(run.SectionGVR.String(), run.FQN, code) {
		c.addIssue(run.FQN, New(run.GroupGVR, run.Group, co.Severity, co.Format(code, args...)))
	}
}

// AddCode add an error code.
func (c *Collector) AddCode(ctx context.Context, code config.ID, args ...interface{}) {
	run := internal.MustExtractRunInfo(ctx)
	co, ok := c.codes.Glossary[code]
	if !ok {
		// BOZO!! refact once codes are in!!
		panic(fmt.Errorf("No code with ID %d", code))
	}
	if !c.ShouldExclude(run.SectionGVR.String(), run.FQN, code) {
		c.addIssue(run.FQN, New(run.SectionGVR, Root, co.Severity, co.Format(code, args...)))
	}
}

// AddErr adds a collection of errors.
func (c *Collector) AddErr(ctx context.Context, errs ...error) {
	run := internal.MustExtractRunInfo(ctx)
	for _, e := range errs {
		c.addIssue(run.FQN, New(run.SectionGVR, Root, config.ErrorLevel, e.Error()))
	}
}

// AddIssue adds 1 or more concerns to the collector.
func (c *Collector) addIssue(fqn string, concerns ...Issue) {
	if len(concerns) == 0 {
		return
	}
	c.outcomes[fqn] = append(c.outcomes[fqn], concerns...)
}
