package cache

import (
	"sync"
	"testing"

	"github.com/magiconair/properties/assert"
	netv1b1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIngressRefs(t *testing.T) {
	ing := NewIngress(map[string]*netv1b1.Ingress{
		"default/ing1": {
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: netv1b1.IngressSpec{
				TLS: []netv1b1.IngressTLS{
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
