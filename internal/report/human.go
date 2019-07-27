package report

import (
	"fmt"
	"strings"
)

func resToTitle() map[string]string {
	return map[string]string{
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
		"psp": "podsecuritypolicie",
	}
}

// Titleize returns a human readable resource name.
func Titleize(r string, count int) string {
	title := r
	if t, ok := resToTitle()[r]; ok {
		title = t
	}

	if count < 0 {
		return strings.ToUpper(fmt.Sprintf("%s", title))
	}
	return strings.ToUpper(fmt.Sprintf("%s (%d scanned)", title+"s", count))
}
