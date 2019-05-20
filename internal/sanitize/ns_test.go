package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespaceSanitizer(t *testing.T) {
	uu := map[string]struct {
		l      NamespaceLister
		issues map[string]int
	}{
		"good": {
			makeNsLister(nsOpts{
				active: true,
				used: []string{
					"ns1",
					"ns2",
					"ns3",
				},
			}),
			map[string]int{"ns1": 0, "ns2": 0, "ns3": 0},
		},
		"inactive": {
			makeNsLister(nsOpts{
				active: false,
				used: []string{
					"ns1",
					"ns2",
					"ns3",
				},
			}),
			map[string]int{"ns1": 0, "ns2": 1, "ns3": 0},
		},
		"unused": {
			makeNsLister(nsOpts{
				active: true,
				used: []string{
					"ns1",
					"ns2",
				},
			}),
			map[string]int{"ns1": 0, "ns2": 0, "ns3": 1},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			n := NewNamespace(issues.NewCollector(), u.l)
			n.Sanitize(context.Background())

			for ns, v := range u.issues {
				assert.Equal(t, v, len(n.Outcome()[ns]))
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type (
	nsOpts struct {
		active bool
		used   []string
	}

	ns struct {
		opts nsOpts
	}
)

func makeNsLister(opts nsOpts) *ns {
	return &ns{
		opts: opts,
	}
}

func (n *ns) ReferencedNamespaces(nn map[string]struct{}) {
	for _, u := range n.opts.used {
		nn[u] = struct{}{}
	}
}

func (n *ns) ListNamespaces() map[string]*v1.Namespace {
	return map[string]*v1.Namespace{
		"ns1": makeNS("ns1", true),
		"ns2": makeNS("ns2", n.opts.active),
		"ns3": makeNS("ns3", true),
	}
}

func makeNS(n string, active bool) *v1.Namespace {
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: n,
		},
	}

	ns.Status.Phase = v1.NamespaceTerminating
	if active {
		ns.Status.Phase = v1.NamespaceActive
	}

	return &ns
}
