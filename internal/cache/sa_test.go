package cache

import (
	"sync"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceAccountRefs(t *testing.T) {
	uu := []struct {
		keys []string
	}{
		{[]string{"sec:default/s1", "sec:default/is1"}},
	}

	sa := NewServiceAccount(map[string]*v1.ServiceAccount{
		"default/sa1": makeSASecrets("sa1"),
	})
	for _, u := range uu {
		var refs sync.Map
		sa.ServiceAccountRefs(&refs)
		for _, k := range u.keys {
			v, ok := refs.Load(k)
			assert.True(t, ok)
			assert.Equal(t, internal.AllKeys, v.(internal.StringSet))
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeSASecrets(n string) *v1.ServiceAccount {
	sa := makeSA(n)
	sa.Secrets = []v1.ObjectReference{
		{
			Kind:      "Secret",
			Name:      "s1",
			Namespace: "default",
		},
	}
	sa.ImagePullSecrets = []v1.LocalObjectReference{
		{
			Name: "is1",
		},
	}

	return sa
}

func makeSA(n string) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
	}
}
