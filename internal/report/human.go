package report

import (
	"fmt"
	"strings"
)

func human() map[string]string {
	return map[string]string{
		"ing": "ingress",
		"po":  "pod",
		"svc": "service",
		"no":  "node",
		"ns":  "namespace",
		"sa":  "serviceaccount",
		"cm":  "configmap",
		"sec": "secret",
		"pv":  "persistentvolume",
		"pvc": "persistentvolumeclaim",
		"hpa": "horizontalpodautoscaler",
		"dp":  "deployment",
		"ds":  "daemonset",
		"sts": "statefulset",
		"pdb": "poddisruptionbudget",
		"np":  "networkpolicy",
		"psp": "podsecuritypolicy",
		"rs":  "replicaset",
		"cl":  "cluster",
	}
}

// ResToTitle converts a resource name to a title if any.
func ResToTitle(r string) string {
	title := r
	if t, ok := human()[r]; ok {
		title = t
	}

	return title
}

// Titleize returns a human readable resource name.
func Titleize(r string, count int) string {
	title := r
	if t, ok := human()[r]; ok {
		title = t
	}

	if count <= 0 || title == "general" {
		return strings.ToUpper(fmt.Sprintf("%s", title))
	}
	return strings.ToUpper(fmt.Sprintf("%s (%d scanned)", title+"s", count))
}
