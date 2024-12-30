// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package issues_test

import (
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/rules"
	"github.com/stretchr/testify/assert"
)

func TestCodesLoad(t *testing.T) {
	cc, err := issues.LoadCodes()

	assert.Nil(t, err)
	assert.Equal(t, 117, len(cc.Glossary))
	assert.Equal(t, "No liveness probe", cc.Glossary[103].Message)
	assert.Equal(t, rules.WarnLevel, cc.Glossary[103].Severity)
}

func TestRefine(t *testing.T) {
	cc, err := issues.LoadCodes()
	assert.Nil(t, err)

	ov := rules.Overrides{
		rules.CodeOverride{
			ID:       0,
			Message:  "blah",
			Severity: rules.InfoLevel,
		},

		rules.CodeOverride{
			ID:       100,
			Message:  "blah",
			Severity: rules.InfoLevel,
		},

		rules.CodeOverride{
			ID:       101,
			Message:  "blah",
			Severity: 1000,
		},
	}
	cc.Refine(ov)

	assert.Equal(t, rules.InfoLevel, cc.Glossary[100].Severity)
	assert.Equal(t, rules.WarnLevel, cc.Glossary[101].Severity)
}
