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
