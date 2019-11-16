package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRSSanitize(t *testing.T) {
	uu := map[string]struct {
		lister ReplicaSetLister
		issues issues.Issues
	}{
		"good": {
			lister: makeRSLister("rs", rsOpts{
				rev: "apps/v1",
			}),
			issues: issues.Issues{},
		},
		"deprecated": {
			lister: makeRSLister("rs", rsOpts{
				rev: "extensions/v1beta1",
			}),
			issues: issues.Issues{
				issues.Issue{
					Group:   "__root__",
					Level:   2,
					Message: `[POP-403] Deprecated ReplicaSet API group "extensions/v1beta1". Use "apps/v1" instead`},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			rs := NewReplicaSet(issues.NewCollector(loadCodes(t)), u.lister)

			assert.Nil(t, rs.Sanitize(context.TODO()))
			assert.Equal(t, u.issues, rs.Outcome()["default/rs"])
		})
	}
}

type (
	rsOpts struct {
		rev string
	}

	rs struct {
		name string
		opts rsOpts
	}
)

func makeRSLister(n string, opts rsOpts) *rs {
	return &rs{
		name: n,
		opts: opts,
	}
}

func (r *rs) ListReplicaSets() map[string]*appsv1.ReplicaSet {
	return map[string]*appsv1.ReplicaSet{
		cache.FQN("default", r.name): makeRS(r.name, r.opts),
	}
}

func makeRS(n string, o rsOpts) *appsv1.ReplicaSet {
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
			SelfLink:  "/api/" + o.rev,
		},
		Spec: appsv1.ReplicaSetSpec{},
	}
}
