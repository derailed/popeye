// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFQN(t *testing.T) {
	uu := []struct {
		ns, n string
		e     string
	}{
		{"", "fred", "fred"},
		{"blee", "fred", "blee/fred"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, FQN(u.ns, u.n))
	}
}

func TestMetaFQN(t *testing.T) {
	uu := []struct {
		m metav1.ObjectMeta
		e string
	}{
		{metav1.ObjectMeta{Namespace: "", Name: "fred"}, "fred"},
		{metav1.ObjectMeta{Namespace: "blee", Name: "fred"}, "blee/fred"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, MetaFQN(u.m))
	}
}

func TestMatchLabels(t *testing.T) {
	uu := map[string]struct {
		labels, selector map[string]string
		e                bool
	}{
		"empty": {
			map[string]string{},
			map[string]string{},
			false,
		},
		"wrongKey": {
			map[string]string{"a": "a", "b": "b"},
			map[string]string{"c": "a"},
			false,
		},
		"wrongValue": {
			map[string]string{"a": "a", "b": "b"},
			map[string]string{"a": "z"},
			false,
		},
		"exact": {
			map[string]string{"a": "a", "b": "b"},
			map[string]string{"a": "a", "b": "b"},
			true,
		},
		"subset": {
			map[string]string{"a": "a", "b": "b", "c": "c"},
			map[string]string{"a": "a", "b": "b"},
			true,
		},
		"superset": {
			map[string]string{"a": "a", "b": "b"},
			map[string]string{"a": "a", "b": "b", "c": "c"},
			false,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, MatchLabels(u.labels, u.selector))
		})
	}
}
