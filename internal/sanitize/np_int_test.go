// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package sanitize

import (
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCheckPodSelector(t *testing.T) {
	uu := map[string]struct {
		nss    map[string]*v1.Namespace
		sel    *metav1.LabelSelector
		issues issues.Issues
	}{
		"empty": {
			nss: map[string]*v1.Namespace{
				"ns1": {ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
			},
		},
		"duh": {
			nss: map[string]*v1.Namespace{
				"ns1": {ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
			},
			sel: &metav1.LabelSelector{MatchLabels: map[string]string{"fred": "blee"}},
			issues: issues.Issues{
				issues.Issue{
					GVR:     "networking.k8s.io/v1/networkpolicies",
					Group:   "__root__",
					Level:   2,
					Message: "[POP-1200] No pods match Ingress pod selector",
				},
			},
		},
	}

	l := makeNPLister(npOpts{
		rev: "networking.k8s.io/v1",
		pod: true,
	})
	np := NewNetworkPolicy(issues.NewCollector(loadCodes(t), makeConfig(t)), l)
	ctx := makeContext("networking.k8s.io/v1/networkpolicies", "np")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			np.checkPodSelector(ctx, u.nss, u.sel, "Ingress")
			assert.Equal(t, u.issues, np.Outcome()[""])
		})

	}
}
