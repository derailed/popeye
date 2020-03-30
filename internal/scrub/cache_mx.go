package scrub

import (
	"sync"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
)

type mx struct {
	*dial

	mx     sync.Mutex
	nodeMx *cache.NodesMetrics
	podMx  *cache.PodsMetrics
}

func newMX(d *dial) *mx {
	return &mx{dial: d}
}

func (m *mx) podsMx() (*cache.PodsMetrics, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if m.podMx != nil {
		return m.podMx, nil
	}
	pmx, err := dag.ListPodsMetrics(m.factory.Client())
	m.podMx = cache.NewPodsMetrics(pmx)

	return m.podMx, err
}

func (m *mx) nodesMx() (*cache.NodesMetrics, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if m.nodeMx != nil {
		return m.nodeMx, nil
	}
	nmx, err := dag.ListNodesMetrics(m.factory.Client())
	m.nodeMx = cache.NewNodesMetrics(nmx)

	return m.nodeMx, err
}
