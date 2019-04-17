package linter

import (
	"fmt"
	"math"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fqn(ns, n string) string {
	return ns + "/" + n
}

func metaFQN(m metav1.ObjectMeta) string {
	if m.Namespace == "" {
		return m.Name
	}
	return fqn(m.Namespace, m.Name)
}

// ToPerc computes the percentage from one number over another.
func ToPerc(v1, v2 int64) int64 {
	if v2 == 0 {
		return 0
	}
	return int64((float64(v1) / float64(v2)) * 100)
}

// In checks if a string is in a list of strings.
func in(ll []string, s string) bool {
	for _, l := range ll {
		if l == s {
			return true
		}
	}

	return false
}

func toMC(q resource.Quantity) int64 {
	return q.MilliValue()
}

const megaByte = 1024 * 1024

func toMB(q resource.Quantity) int64 {
	return q.Value() / megaByte
}

func asPerc(n int64) string {
	return fmt.Sprintf("%d%%", n)
}

func toRatio(q1, q2 resource.Quantity) int64 {
	if q2.IsZero() {
		return 0
	}
	return int64(math.Round((float64(q1.MilliValue()) / (float64(q2.MilliValue()))) * 100))
}

func asMC(q resource.Quantity) string {
	return fmt.Sprintf("%vm", toMC(q))
}

func asMB(q resource.Quantity) string {
	return fmt.Sprintf("%vMi", toMB(q))
}
