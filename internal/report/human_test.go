package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitleize(t *testing.T) {
	uu := map[string]string{
		"po":   "PODS (1 SCANNED)",
		"no":   "NODES (1 SCANNED)",
		"svc":  "SERVICES (1 SCANNED)",
		"blee": "BLEES (1 SCANNED)",
		"sa":   "SERVICEACCOUNTS (1 SCANNED)",
		"ns":   "NAMESPACES (1 SCANNED)",
		"cm":   "CONFIGMAPS (1 SCANNED)",
		"sec":  "SECRETS (1 SCANNED)",
		"pv":   "PERSISTENTVOLUMES (1 SCANNED)",
		"pvc":  "PERSISTENTVOLUMECLAIMS (1 SCANNED)",
		"hpa":  "HORIZONTALPODAUTOSCALERS (1 SCANNED)",
		"dp":   "DEPLOYMENTS (1 SCANNED)",
		"sts":  "STATEFULSETS (1 SCANNED)",
	}

	for k, e := range uu {
		assert.Equal(t, e, Titleize(k, 1))
	}
}
