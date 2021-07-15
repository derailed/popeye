package sanitize

import (
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	polv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPSPSanitize(t *testing.T) {
	uu := map[string]struct {
		lister PodSecurityPolicyLister
		issues issues.Issues
	}{
		"good": {
			lister: makePSPLister("psp", pspOpts{
				rev: "policy/v1beta1",
			}),
			issues: issues.Issues{},
		},
		"deprecated": {
			lister: makePSPLister("psp", pspOpts{
				rev: "extensions/v1beta1",
			}),
			issues: issues.Issues{
				issues.Issue{
					GVR:     "policy/v1beta1/podsecuritypolicies",
					Group:   "__root__",
					Level:   2,
					Message: `[POP-403] Deprecated PodSecurityPolicy API group "extensions/v1beta1". Use "policy/v1beta1" instead`},
			},
		},
	}

	ctx := makeContext("policy/v1beta1/podsecuritypolicies", "psp")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			psp := NewPodSecurityPolicy(issues.NewCollector(loadCodes(t), makeConfig(t)), u.lister)

			assert.Nil(t, psp.Sanitize(ctx))
			assert.Equal(t, u.issues, psp.Outcome()["default/psp"])
		})
	}
}

// Helpers...

type (
	pspOpts struct {
		rev string
	}

	psp struct {
		name string
		opts pspOpts
	}
)

func makePSPLister(n string, opts pspOpts) *psp {
	return &psp{
		name: n,
		opts: opts,
	}
}

func (r *psp) ListPodSecurityPolicies() map[string]*polv1beta1.PodSecurityPolicy {
	return map[string]*polv1beta1.PodSecurityPolicy{
		cache.FQN("default", r.name): makePSP(r.name, r.opts),
	}
}

func makePSP(n string, o pspOpts) *polv1beta1.PodSecurityPolicy {
	return &polv1beta1.PodSecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
			SelfLink:  "/api/" + o.rev,
		},
		Spec: polv1beta1.PodSecurityPolicySpec{},
	}
}
