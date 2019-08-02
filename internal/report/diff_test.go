package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiff(t *testing.T) {
	r1, r2 := "r1.yml", "r2.yml"
	b1, err := loadBuilder("assets", r1)
	assert.Nil(t, err)

	b2, err := loadBuilder("assets", r2)
	assert.Nil(t, err)

	diff := newDiffReport(&b1.Report, &b2.Report)
	diff.Build()

	assert.Equal(t, "-50", diff.overall.delta())
	assert.Equal(t, 1, len(diff.sections))
}
