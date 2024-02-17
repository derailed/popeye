// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"fmt"
	"strconv"
	"strings"
)

// Code represents an issue code.
type Code struct {
	Message  string `yaml:"message"`
	Severity Level  `yaml:"severity"`
}

// Format hydrates a message with arguments.
func (c *Code) Format(code ID, args ...any) string {
	msg := "[POP-" + strconv.Itoa(int(code)) + "] "
	if len(args) == 0 {
		msg += c.Message
	} else {
		msg += fmt.Sprintf(c.Message, args...)
	}

	return strings.TrimSpace(msg)
}
