// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestLimitCPURatio(t *testing.T) {
	uu := map[string]struct {
		mx ConsumptionMetrics
		e  float64
	}{
		"empty": {},
		"same": {
			mx: ConsumptionMetrics{
				CurrentCPU: *resource.NewQuantity(10, resource.DecimalExponent),
				LimitCPU:   *resource.NewQuantity(10, resource.DecimalExponent),
			},
			e: 100,
		},
		"delta": {
			mx: ConsumptionMetrics{
				CurrentCPU: *resource.NewQuantity(100, resource.DecimalExponent),
				LimitCPU:   *resource.NewQuantity(10, resource.DecimalExponent),
			},
			e: 1_000,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.mx.LimitCPURatio())
		})
	}
}

func TestReqAbsCPURatio(t *testing.T) {
	uu := map[string]struct {
		mx ConsumptionMetrics
		e  float64
	}{
		"empty": {},
		"same": {
			mx: ConsumptionMetrics{
				CurrentCPU: *resource.NewQuantity(10, resource.DecimalExponent),
				RequestCPU: *resource.NewQuantity(10, resource.DecimalExponent),
			},
			e: 100,
		},
		"higher": {
			mx: ConsumptionMetrics{
				CurrentCPU: *resource.NewQuantity(2, resource.DecimalExponent),
				RequestCPU: *resource.NewQuantity(10, resource.DecimalExponent),
			},
			e: 500,
		},
		"lower": {
			mx: ConsumptionMetrics{
				CurrentCPU: *resource.NewQuantity(10, resource.DecimalExponent),
				RequestCPU: *resource.NewQuantity(100, resource.DecimalExponent),
			},
			e: 1_000,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.mx.ReqAbsCPURatio())
		})
	}
}

func TestReqCPURatio(t *testing.T) {
	uu := map[string]struct {
		mx ConsumptionMetrics
		e  float64
	}{
		"empty": {},
		"same": {
			mx: ConsumptionMetrics{
				CurrentCPU: *resource.NewQuantity(10, resource.DecimalExponent),
				RequestCPU: *resource.NewQuantity(10, resource.DecimalExponent),
			},
			e: 100,
		},
		"higher": {
			mx: ConsumptionMetrics{
				CurrentCPU: *resource.NewQuantity(100, resource.DecimalExponent),
				RequestCPU: *resource.NewQuantity(10, resource.DecimalExponent),
			},
			e: 1000,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.mx.ReqCPURatio())
		})
	}
}

func TestReqAbsMEMRatio(t *testing.T) {
	uu := map[string]struct {
		mx ConsumptionMetrics
		e  float64
	}{
		"empty": {},
		"same": {
			mx: ConsumptionMetrics{
				CurrentMEM: *resource.NewQuantity(10*megaByte, resource.DecimalExponent),
				RequestMEM: *resource.NewQuantity(10*megaByte, resource.DecimalExponent),
			},
			e: 100,
		},
		"higher": {
			mx: ConsumptionMetrics{
				CurrentMEM: *resource.NewQuantity(100*megaByte, resource.DecimalExponent),
				RequestMEM: *resource.NewQuantity(10*megaByte, resource.DecimalExponent),
			},
			e: 1_000,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.mx.ReqAbsMEMRatio())
		})
	}
}

func TestReqMEMRatio(t *testing.T) {
	uu := map[string]struct {
		mx ConsumptionMetrics
		e  float64
	}{
		"empty": {},
		"same": {
			mx: ConsumptionMetrics{
				CurrentMEM: *resource.NewQuantity(10*megaByte, resource.DecimalExponent),
				RequestMEM: *resource.NewQuantity(10*megaByte, resource.DecimalExponent),
			},
			e: 100,
		},
		"delta": {
			mx: ConsumptionMetrics{
				CurrentMEM: *resource.NewQuantity(100*megaByte, resource.DecimalExponent),
				RequestMEM: *resource.NewQuantity(10*megaByte, resource.DecimalExponent),
			},
			e: 1_000,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.mx.ReqMEMRatio())
		})
	}
}

func TestLimitMEMRatio(t *testing.T) {
	uu := map[string]struct {
		mx ConsumptionMetrics
		e  float64
	}{
		"empty": {},
		"same": {
			mx: ConsumptionMetrics{
				CurrentMEM: *resource.NewQuantity(10*megaByte, resource.DecimalExponent),
				LimitMEM:   *resource.NewQuantity(10*megaByte, resource.DecimalExponent),
			},
			e: 100,
		},
		"delta": {
			mx: ConsumptionMetrics{
				CurrentMEM: *resource.NewQuantity(100*megaByte, resource.DecimalExponent),
				LimitMEM:   *resource.NewQuantity(10*megaByte, resource.DecimalExponent),
			},
			e: 1_000,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.mx.LimitMEMRatio())
		})
	}
}
