package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSecretSanitize(t *testing.T) {
	s := NewSecret(issues.NewCollector(loadCodes(t)), newSecret())

	assert.Nil(t, s.Sanitize(context.TODO()))
	assert.Equal(t, 5, len(s.Outcome()))

	ii := s.Outcome()["default/sec3"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, "[POP-400] Used? Unable to locate resource reference", ii[0].Message)
	assert.Equal(t, issues.InfoLevel, ii[0].Level)

	ii = s.Outcome()["default/sec4"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-401] Key "k2" used? Unable to locate key reference`, ii[0].Message)
	assert.Equal(t, issues.InfoLevel, ii[0].Level)
}

// ----------------------------------------------------------------------------
// Helpers...

type secretMock struct{}

func newSecret() secretMock {
	return secretMock{}
}

func (m secretMock) PodRefs(refs cache.ObjReferences) {
	refs["sec:default/sec1"] = cache.StringSet{
		"k1": cache.Blank,
		"k2": cache.Blank,
	}
	refs["sec:default/sec2"] = cache.StringSet{
		"all": cache.Blank,
	}
	refs["sec:default/sec4"] = cache.StringSet{
		"k1": cache.Blank,
	}
}

func (m secretMock) IngressRefs(cache.ObjReferences) {}

func (m secretMock) ServiceAccountRefs(refs cache.ObjReferences) {
	refs["sec:default/sec5"] = cache.StringSet{
		"all": cache.Blank,
	}
}

func (m secretMock) ListSecrets() map[string]*v1.Secret {
	return map[string]*v1.Secret{
		"default/sec1": makeSecret("sec1"),
		"default/sec2": makeSecret("sec2"),
		"default/sec3": makeSecret("sec3"),
		"default/sec4": makeSecret("sec4"),
		"default/sec5": makeSecret("sec5"),
	}
}

func makeSecret(n string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Data: map[string][]byte{
			"k1": {},
			"k2": {},
		},
	}
}

func makeDockerSecret(n string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Type: v1.SecretTypeDockercfg,
		Data: map[string][]byte{
			"k1": {},
			"k2": {},
		},
	}
}
