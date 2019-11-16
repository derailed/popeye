package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitleize(t *testing.T) {
	uu := map[string][]string{
		"cl":  {"CLUSTER", "CLUSTERS (1 SCANNED)"},
		"cm":  {"CONFIGMAP", "CONFIGMAPS (1 SCANNED)"},
		"dp":  {"DEPLOYMENT", "DEPLOYMENTS (1 SCANNED)"},
		"ds":  {"DAEMONSET", "DAEMONSETS (1 SCANNED)"},
		"hpa": {"HORIZONTALPODAUTOSCALER", "HORIZONTALPODAUTOSCALERS (1 SCANNED)"},
		"ing": {"INGRESS", "INGRESSES (1 SCANNED)"},
		"no":  {"NODE", "NODES (1 SCANNED)"},
		"np":  {"NETWORKPOLICY", "NETWORKPOLICIES (1 SCANNED)"},
		"ns":  {"NAMESPACE", "NAMESPACES (1 SCANNED)"},
		"pdb": {"PODDISRUPTIONBUDGET", "PODDISRUPTIONBUDGETS (1 SCANNED)"},
		"po":  {"POD", "PODS (1 SCANNED)"},
		"psp": {"PODSECURITYPOLICY", "PODSECURITYPOLICIES (1 SCANNED)"},
		"pv":  {"PERSISTENTVOLUME", "PERSISTENTVOLUMES (1 SCANNED)"},
		"pvc": {"PERSISTENTVOLUMECLAIM", "PERSISTENTVOLUMECLAIMS (1 SCANNED)"},
		"rs":  {"REPLICASET", "REPLICASETS (1 SCANNED)"},
		"sa":  {"SERVICEACCOUNT", "SERVICEACCOUNTS (1 SCANNED)"},
		"sec": {"SECRET", "SECRETS (1 SCANNED)"},
		"sts": {"STATEFULSET", "STATEFULSETS (1 SCANNED)"},
		"svc": {"SERVICE", "SERVICES (1 SCANNED)"},

		// Fallback
		"blee": {"BLEE", "BLEES (1 SCANNED)"},
	}

	for k, e := range uu {
		assert.Equal(t, e[0], Titleize(k, 0))
		assert.Equal(t, e[1], Titleize(k, 1))
	}
}
