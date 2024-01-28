// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FQN returns a fully qualified resource identifier.
func FQN(ns, n string) string {
	if ns == "" {
		return n
	}
	return ns + "/" + n
}

// MetaFQN returns a fully qualified resource identifier based on object meta.
func MetaFQN(m metav1.ObjectMeta) string {
	return FQN(m.Namespace, m.Name)
}

// ResFqn returns a resource specific fqn.
func ResFqn(k, s string) string {
	return k + ":" + s
}

// Namespaced return ns and name contained in given fqn.
func namespaced(fqn string) (string, string) {
	tokens := strings.Split(fqn, "/")
	if len(tokens) == 2 {
		return tokens[0], tokens[1]
	}
	return "", tokens[0]
}

// MatchLabels check if pod labels match a selector.
func MatchLabels(labels, sel map[string]string) bool {
	if len(sel) == 0 {
		return false
	}

	for k, v := range sel {
		if v1, ok := labels[k]; !ok || v1 != v {
			return false
		}
	}

	return true
}
