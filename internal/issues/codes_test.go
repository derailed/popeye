package issues_test

import (
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
)

func TestCodesLoad(t *testing.T) {
	cc, err := issues.LoadCodes("./assets/codes.yml")

	assert.Nil(t, err)
	assert.Equal(t, 13, len(cc.Glossary))
	assert.Equal(t, "No liveness probe", cc.Glossary[103].Message)
	assert.Equal(t, issues.WarnLevel, cc.Glossary[103].Severity)
}

func TestRefine(t *testing.T) {
	cc, err := issues.LoadCodes("./assets/codes.yml")
	assert.Nil(t, err)

	id1, id2 := issues.ID(100), issues.ID(101)
	gloss := issues.Glossary{
		0: &issues.Code{
			Message:  "blah",
			Severity: issues.InfoLevel,
		},

		id1: &issues.Code{
			Message:  "blah",
			Severity: issues.InfoLevel,
		},
		id2: &issues.Code{
			Message:  "blah",
			Severity: 1000,
		},
	}
	cc.Refine(gloss)

	assert.Equal(t, issues.InfoLevel, cc.Glossary[id1].Severity)
	assert.Equal(t, issues.WarnLevel, cc.Glossary[id2].Severity)
}
