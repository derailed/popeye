package issues

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"gopkg.in/yaml.v2"
)

type (
	// ID represents a sanitizer code indentifier.
	ID int

	// Glossary represents a collection of codes.
	Glossary map[ID]*Code

	// Codes represents a collection of sanitizer codes.
	Codes struct {
		Glossary Glossary `yaml:"codes"`
	}

	// Code represents a sanitizer code.
	Code struct {
		Message  string `yaml:"message"`
		Severity Level  `yaml:"severity"`
	}
)

// LoadCodes retrieves sanitifizers codes from yaml file.
func LoadCodes(path string) (*Codes, error) {
	var cc Codes
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return &cc, err
	}
	if err := yaml.Unmarshal(raw, &cc); err != nil {
		return &cc, err
	}

	return &cc, err
}

// Refine overrides code severity based on user input.
func (c *Codes) Refine(gloss Glossary) {
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

// Format hydrates a message with arguments.
func (c *Code) Format(code ID, args ...interface{}) string {
	msg := "[POP-" + strconv.Itoa(int(code)) + "] "
	if len(args) == 0 {
		return msg + c.Message
	}
	return msg + fmt.Sprintf(c.Message, args...)
}

// Helpers...

func validSeverity(l Level) bool {
	return l > 0 && l < 4
}
