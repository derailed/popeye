package scrub

import (
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
)

type mx struct {
	*dial

	nodeMx *cache.NodesMetrics
	podMx  *cache.PodsMetrics
}

func newMX(d *dial) *mx {
	return &mx{dial: d}
}

func (m *mx) podsMx() (*cache.PodsMetrics, error) {
	if m.podMx != nil {
		return m.podMx, nil
	}
	pmx, err := dag.ListPodsMetrics(m.factory.Client())
	m.podMx = cache.NewPodsMetrics(pmx)

	return m.podMx, err
}

func (m *mx) nodesMx() (*cache.NodesMetrics, error) {
	if m.nodeMx != nil {
		return m.nodeMx, nil
	}
	nmx, err := dag.ListNodesMetrics(m.factory.Client())
	m.nodeMx = cache.NewNodesMetrics(nmx)

	return m.nodeMx, err
}
