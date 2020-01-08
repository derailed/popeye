package sanitize

import (
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
	nv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIngressSanitize(t *testing.T) {
	uu := map[string]struct {
		rev string
		e   issues.Issues
	}{
		"good": {
			rev: "networking.k8s.io/v1beta1",
			e:   issues.Issues{},
		},
		"guizard": {
			rev: "extensions/v1beta1",
			e: issues.Issues{
				{
					Group:   issues.Root,
					Message: `[POP-403] Deprecated Ingress API group "extensions/v1beta1". Use "networking.k8s.io/v1beta1" instead`,
					Level:   config.WarnLevel,
				},
			},
		},
	}

	ctx := makeContext("ing")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cl := NewIngress(issues.NewCollector(loadCodes(t), makeConfig(t)), newIngress(u.rev))

			assert.Nil(t, cl.Sanitize(ctx))
			assert.Equal(t, u.e, cl.Outcome()["default/ing1"])
		})
	}
}

// Helpers...

type ingress struct {
	rev string
}

func newIngress(rev string) ingress {
	return ingress{rev: rev}
}

func (i ingress) ListIngresses() map[string]*nv1beta1.Ingress {
	return map[string]*nv1beta1.Ingress{
		"default/ing1": makeIngress(i.rev),
	}
}

func makeIngress(url string) *nv1beta1.Ingress {
	return &nv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			SelfLink: "/api/" + url,
		},
	}
}
