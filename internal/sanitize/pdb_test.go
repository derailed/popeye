package sanitize

import (
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestPDBSanitize(t *testing.T) {
	uu := map[string]struct {
		lister PodDisruptionBudgetLister
		issues issues.Issues
	}{
		"good": {
			lister: makePDBLister(pdbOpts{}),
			issues: issues.Issues{},
		},
		"noPods": {
			lister: makePDBLister(pdbOpts{pod: true}),
			issues: issues.Issues{
				issues.Issue{
					GVR:     "policy/v1beta1/poddisruptionbudgets",
					Group:   "__root__",
					Level:   2,
					Message: "[POP-900] Used? No pods match selector"},
			},
		},
	}

	ctx := makeContext("policy/v1beta1/poddisruptionbudgets", "pdb")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pdb := NewPodDisruptionBudget(issues.NewCollector(loadCodes(t), makeConfig(t)), u.lister)

			assert.Nil(t, pdb.Sanitize(ctx))
			assert.Equal(t, u.issues, pdb.Outcome()["default/pdb"])
		})
	}
}

// Helpers...

type (
	pdbOpts struct {
		pod bool
	}

	pdb struct {
		name string
		opts pdbOpts
	}
)

func makePDBLister(opts pdbOpts) *pdb {
	return &pdb{
		name: "pdb",
		opts: opts,
	}
}

func (r *pdb) ListPodDisruptionBudgets() map[string]*policyv1.PodDisruptionBudget {
	return map[string]*policyv1.PodDisruptionBudget{
		cache.FQN("default", r.name): makePDB(r.name),
	}
}

func (r *pdb) ListPods() map[string]*v1.Pod {
	return map[string]*v1.Pod{
		"default/p1": makePodSa("p1", "fred"),
	}
}

func (r *pdb) GetPod(ns string, sel map[string]string) *v1.Pod {
	if r.opts.pod {
		return nil
	}
	return makePod("p1")
}

func makePDB(n string) *policyv1.PodDisruptionBudget {
	min, max := intstr.FromInt(1), intstr.FromInt(1)
	return &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector:       &metav1.LabelSelector{},
			MinAvailable:   &min,
			MaxUnavailable: &max,
		},
	}
}
