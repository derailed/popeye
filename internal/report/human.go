package report

import (
	"fmt"
	"strings"
)

func human() map[string][]string {
	return map[string][]string{
		"cl":  {"cluster"},
		"cm":  {"configmap"},
		"dp":  {"deployment"},
		"ds":  {"daemonset"},
		"hpa": {"horizontalpodautoscaler"},
		"ing": {"ingress", "ingresses"},
		"no":  {"node"},
		"np":  {"networkpolicy", "networkpolicies"},
		"ns":  {"namespace"},
		"pdb": {"poddisruptionbudget"},
		"po":  {"pod"},
		"psp": {"podsecuritypolicy", "podsecuritypolicies"},
		"pv":  {"persistentvolume"},
		"pvc": {"persistentvolumeclaim"},
		"rs":  {"replicaset"},
		"sa":  {"serviceaccount"},
		"sec": {"secret"},
		"sts": {"statefulset"},
		"svc": {"service"},
	}
}

// ResToTitle converts a resource name to a title if any.
func ResToTitle(r string) string {
	title := r
	if t, ok := human()[r]; ok {
		title = t[0]
	}

	return title
}

// Titleize returns a human readable resource name.
func Titleize(r string, count int) string {
	inflections := inflectResourceWord(r)

	title := inflections[0]
	if count <= 0 || title == "general" {
		return strings.ToUpper(fmt.Sprintf("%s", title))
	}

	title = inflections[1]
	return strings.ToUpper(fmt.Sprintf("%s (%d scanned)", title, count))
}

func inflectResourceWord(r string) []string {
	inflections := []string{r}
	if i, ok := human()[r]; ok {
		inflections = i
	}
	if len(inflections) == 1 {
		inflections = []string{inflections[0], inflections[0] + "s"}
	}
	return inflections
}
