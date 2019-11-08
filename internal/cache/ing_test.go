package cache

import (
	"testing"

	"github.com/magiconair/properties/assert"
	nv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIngressRefs(t *testing.T) {
	ing := NewIngress(map[string]*nv1beta1.Ingress{
		"default/ing1": &nv1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: nv1beta1.IngressSpec{
				TLS: []nv1beta1.IngressTLS{
					{
						SecretName: "foo",
					},
				},
			},
		},
	})

	refs := ObjReferences{}
	ing.IngressRefs(refs)

	_, exists := refs["sec:default/foo"]

	assert.Equal(t, len(refs), 1)
	assert.Equal(t, exists, true)
}
