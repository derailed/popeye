package report

import (
	"fmt"
	"strings"
)

// Titleize returns a human readable resource name.
func Titleize(r string, count int) string {
	var t string
	switch r {
	case "po":
		t = "pod"
	case "svc":
		t = "service"
	case "no":
		t = "node"
	case "ns":
		t = "namespace"
	case "sa":
		t = "serviceaccount"
	case "cm":
		t = "configmap"
	case "sec":
		t = "Secret"
	case "pv":
		t = "persistentvolume"
	case "pvc":
		t = "persistentvolumeclaim"
	case "hpa":
		t = "horizontalpodautoscaler"
	case "dp":
		t = "deployment"
	case "sts":
		t = "statefulset"
	default:
		t = r
	}

	if count < 0 {
		return strings.ToUpper(fmt.Sprintf("%s", t))
	}
	return strings.ToUpper(fmt.Sprintf("%s (%d scanned)", t+"s", count))
}
