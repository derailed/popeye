package issues_test

import (
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestCodesLoad(t *testing.T) {
	cc, err := issues.LoadCodes()

	assert.Nil(t, err)
	assert.Equal(t, 85, len(cc.Glossary))
	assert.Equal(t, "No liveness probe", cc.Glossary[103].Message)
	assert.Equal(t, config.WarnLevel, cc.Glossary[103].Severity)
}

func TestRefine(t *testing.T) {
	cc, err := issues.LoadCodes()
	assert.Nil(t, err)

	id1, id2 := config.ID(100), config.ID(101)
	gloss := config.Glossary{
		0: &config.Code{
			Message:  "blah",
			Severity: config.InfoLevel,
		},

		id1: &config.Code{
			Message:  "blah",
			Severity: config.InfoLevel,
		},
		id2: &config.Code{
			Message:  "blah",
			Severity: 1000,
		},
	}
	cc.Refine(gloss)

	assert.Equal(t, config.InfoLevel, cc.Glossary[id1].Severity)
	assert.Equal(t, config.WarnLevel, cc.Glossary[id2].Severity)
}
