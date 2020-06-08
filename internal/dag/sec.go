package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListSecrets list all included Secrets.
func ListSecrets(ctx context.Context) (map[string]*v1.Secret, error) {
	return listAllSecrets(ctx)
}

// ListAllSecrets fetch all Secrets on the cluster.
func listAllSecrets(ctx context.Context) (map[string]*v1.Secret, error) {
	ll, err := fetchSecrets(ctx)
	if err != nil {
		return nil, err
	}
	secs := make(map[string]*v1.Secret, len(ll.Items))
	for i := range ll.Items {
		secs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return secs, nil
}

// FetchSecrets retrieves all Secrets on the cluster.
func fetchSecrets(ctx context.Context) (*v1.SecretList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.CoreV1().Secrets(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("v1/secrets"))
	oo, err := res.List(ctx)
	if err != nil {
		return nil, err
	}
	var ll v1.SecretList
	for _, o := range oo {
		var sec v1.Secret
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &sec)
		if err != nil {
			return nil, errors.New("expecting secret resource")
		}
		ll.Items = append(ll.Items, sec)
	}

	return &ll, nil
}
