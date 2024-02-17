// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package issues

import (
	_ "embed"

	"github.com/derailed/popeye/internal/rules"
	"gopkg.in/yaml.v2"
)

//go:embed assets/codes.yaml
var codes string

// Codes represents a collection of linter codes.
type Codes struct {
	Glossary rules.Glossary `yaml:"codes"`
}

// LoadCodes retrieves linters codes from yaml file.
func LoadCodes() (*Codes, error) {
	var cc Codes
	if err := yaml.Unmarshal([]byte(codes), &cc); err != nil {
		return &cc, err
	}

	return &cc, nil
}

// Refine overrides code severity based on user input.
func (c *Codes) Refine(oo rules.Overrides) {
	for _, ov := range oo {
		c, ok := c.Glossary[ov.ID]
		if !ok {
			continue
		}
		if validSeverity(ov.Severity) {
			c.Severity = ov.Severity
		}
		if ov.Message != "" {
			c.Message = ov.Message
		}
	}
}

// Helpers...

func validSeverity(l rules.Level) bool {
	return l > 0 && l < 4
}
