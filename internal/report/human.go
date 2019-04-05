package report

import "strings"

// TitleForRes returns a human readable resource name.
func TitleForRes(r string) string {
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
	default:
		t = r
	}
	return strings.ToUpper(t + "s")
}
