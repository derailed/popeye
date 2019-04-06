package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitleForRes(t *testing.T) {
	uu := map[string]string{
		"po":   "PODS",
		"no":   "NODES",
		"svc":  "SERVICES",
		"blee": "BLEES",
		"sa":   "SERVICEACCOUNTS",
		"ns":   "NAMESPACES",
	}

	for k, e := range uu {
		assert.Equal(t, e, TitleForRes(k))
	}
}
