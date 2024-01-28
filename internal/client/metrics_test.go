// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNodeMetricsEmpty(t *testing.T) {
	uu := []struct {
		m NodeMetrics
		e bool
	}{
		{NodeMetrics{}, true},
		{NodeMetrics{CurrentCPU: toQty("100m")}, false},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, u.m.Empty())
	}
}

func TestMetricsEmpty(t *testing.T) {
	uu := []struct {
		m Metrics
		e bool
	}{
		{Metrics{}, true},
		{Metrics{CurrentCPU: toQty("100m")}, false},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, u.m.Empty())
	}
}

// Helpers...

func toQty(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)

	return q
}
