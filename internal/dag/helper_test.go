package dag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMetaFQN(t *testing.T) {
	uu := []struct {
		m metav1.ObjectMeta
		e string
	}{
		{metav1.ObjectMeta{Namespace: "", Name: "fred"}, "fred"},
		{metav1.ObjectMeta{Namespace: "blee", Name: "fred"}, "blee/fred"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, metaFQN(u.m))
	}
}
