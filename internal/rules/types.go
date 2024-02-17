// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import "strconv"

const ZeroCode ID = 0

// ID represents a issue code identifier.
type ID int

func (i ID) String() string {
	return strconv.Itoa(int(i))
}

// IDS tracks a collection of ids.
type IDS map[Code]struct{}

type CodeOverride struct {
	ID       ID     `yaml:"code"`
	Message  string `yaml:"message"`
	Severity Level  `yaml:"severity"`
}

// Glossary represents a collection of codes.
type Overrides []CodeOverride

// Glossary represents a collection of codes.
type Glossary map[ID]*Code
