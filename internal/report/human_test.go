package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitleize(t *testing.T) {
	uu := map[string]string{
		"po":   "PODS",
		"no":   "NODES",
		"svc":  "SERVICES",
		"blee": "BLEES",
		"sa":   "SERVICEACCOUNTS",
		"ns":   "NAMESPACES",
		"cm":   "CONFIGMAPS",
		"sec":  "SECRETS",
		"pv":   "PERSISTENTVOLUMES",
		"pvc":  "PERSISTENTVOLUMECLAIMS",
		"hpa":  "HORIZONTALPODAUTOSCALERS",
		"dp":   "DEPLOYMENTS",
		"sts":  "STATEFULSETS",
	}

	for k, e := range uu {
		assert.Equal(t, e, Titleize(k))
	}
}
