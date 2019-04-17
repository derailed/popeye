package report

import "strings"

// Titleize returns a human readable resource name.
func Titleize(r string) string {
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
	return strings.ToUpper(t + "s")
}
