package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	pv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestPDBSanitize(t *testing.T) {
	uu := map[string]struct {
		lister PodDisruptionBudgetLister
		issues issues.Issues
	}{
		"good": {
			lister: makePDBLister("pdb", pdbOpts{}),
			issues: issues.Issues{},
		},
		"noPods": {
			lister: makePDBLister("pdb", pdbOpts{pod: true}),
			issues: issues.Issues{
				issues.Issue{
					Group:   "__root__",
					Level:   2,
					Message: "[POP-900] Used? No pods match selector"},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			pdb := NewPodDisruptionBudget(issues.NewCollector(loadCodes(t)), u.lister)
			pdb.Sanitize(context.Background())

			assert.Equal(t, u.issues, pdb.Outcome()["default/pdb"])
		})
	}
}

type (
	pdbOpts struct {
		pod bool
	}

	pdb struct {
		name string
		opts pdbOpts
	}
)

func makePDBLister(n string, opts pdbOpts) *pdb {
	return &pdb{
		name: n,
		opts: opts,
	}
}

func (r *pdb) ListPodDisruptionBudgets() map[string]*pv1beta1.PodDisruptionBudget {
	return map[string]*pv1beta1.PodDisruptionBudget{
		cache.FQN("default", r.name): makePDB(r.name, r.opts),
	}
}

func (r *pdb) ListPods() map[string]*v1.Pod {
	return map[string]*v1.Pod{
		"default/p1": makePodSa("p1", "fred"),
	}
}

func (r *pdb) GetPod(map[string]string) *v1.Pod {
	if r.opts.pod {
		return nil
	}
	return makePod("p1")
}

func makePDB(n string, o pdbOpts) *pv1beta1.PodDisruptionBudget {
	min, max := intstr.FromInt(1), intstr.FromInt(1)
	return &pv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Spec: pv1beta1.PodDisruptionBudgetSpec{
			Selector:       &metav1.LabelSelector{},
			MinAvailable:   &min,
			MaxUnavailable: &max,
		},
	}
}
