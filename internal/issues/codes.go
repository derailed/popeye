package issues

import (
	// Pull in asset codes.
	_ "embed"

	"github.com/derailed/popeye/pkg/config"
	"gopkg.in/yaml.v2"
)

type (
	// Codes represents a collection of sanitizer codes.
	Codes struct {
		Glossary config.Glossary `yaml:"codes"`
	}
)

// LoadCodes retrieves sanitizers codes from yaml file.
func LoadCodes() (*Codes, error) {
	var cc Codes
	if err := yaml.Unmarshal([]byte(codes), &cc); err != nil {
		return &cc, err
	}

	return &cc, nil
}

// Refine overrides code severity based on user input.
func (c *Codes) Refine(gloss config.Glossary) {
	for k, v := range gloss {
		c, ok := c.Glossary[k]
		if !ok {
			continue
		}
		if validSeverity(v.Severity) {
			c.Severity = v.Severity
		}
	}
}

// Helpers...

func validSeverity(l config.Level) bool {
	return l > 0 && l < 4
}

//go:embed assets/codes.yml
var codes string
