package linter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type nodeMetrics struct {
	cpu, maxCPU int64
	mem, maxMEM int64
}

func (n nodeMetrics) CurrentCPU() int64 {
	return n.cpu
}
func (n nodeMetrics) MaxCPU() int64 {
	return n.maxCPU
}
func (n nodeMetrics) CurrentMEM() int64 {
	return n.mem
}
func (n nodeMetrics) MaxMEM() int64 {
	return n.maxMEM
}
func (n nodeMetrics) Empty() bool {
	return n.cpu == 0 && n.mem == 0 && n.maxCPU == 0 && n.maxMEM == 0
}

func TestNodeUtilization(t *testing.T) {
	uu := []struct {
		mx     nodeMetrics
		issues int
		level  Level
	}{
		{
			mx:     nodeMetrics{cpu: 500, maxCPU: 1000, mem: 1000, maxMEM: 2000},
			issues: 0,
		},
		{
			mx:     nodeMetrics{cpu: 800, maxCPU: 1000, mem: 1000, maxMEM: 2000},
			issues: 1,
			level:  WarnLevel,
		},
		{
			mx:     nodeMetrics{cpu: 500, maxCPU: 1000, mem: 8000, maxMEM: 10000},
			issues: 1,
			level:  WarnLevel,
		},
		{
			mx:     nodeMetrics{cpu: 900, maxCPU: 1000, mem: 9000, maxMEM: 10000},
			issues: 2,
			level:  WarnLevel,
		},
	}

	for _, u := range uu {
		l := NewNode()
		l.checkUtilization("blee", u.mx)
		assert.Equal(t, u.issues, len(l.Issues()))
	}
}
