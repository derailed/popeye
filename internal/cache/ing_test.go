package cache

import (
	"sync"
	"testing"

	"github.com/magiconair/properties/assert"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIngressRefs(t *testing.T) {
	ing := NewIngress(map[string]*netv1.Ingress{
		"default/ing1": {
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: netv1.IngressSpec{
				TLS: []netv1.IngressTLS{
					{
						SecretName: "foo",
					},
				},
			},
		},
	})

	var refs sync.Map
	ing.IngressRefs(&refs)

	_, ok := refs.Load("sec:default/foo")
	assert.Equal(t, ok, true)
}
