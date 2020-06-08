package dag

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mustExtractFactory(ctx context.Context) types.Factory {
	f, ok := ctx.Value(internal.KeyFactory).(types.Factory)
	if !ok {
		panic("expecting factory in context")
	}
	return f
}

func mustExtractConfig(ctx context.Context) *config.Config {
	cfg, ok := ctx.Value(internal.KeyConfig).(*config.Config)
	if !ok {
		panic("expecting config in context")
	}
	return cfg
}

// MetaFQN returns a full qualified ns/name string.
func metaFQN(m metav1.ObjectMeta) string {
	if m.Namespace == "" {
		return m.Name
	}

	return m.Namespace + "/" + m.Name
}
