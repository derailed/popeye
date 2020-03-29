package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListSecrets list all included Secrets.
func ListSecrets(f types.Factory, cfg *config.Config) (map[string]*v1.Secret, error) {
	secs, err := listAllSecrets(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.Secret, len(secs))
	for fqn, sec := range secs {
		if includeNS(f.Client(), sec.Namespace) {
			res[fqn] = sec
		}
	}

	return res, nil
}

// ListAllSecrets fetch all Secrets on the cluster.
func listAllSecrets(f types.Factory) (map[string]*v1.Secret, error) {
	ll, err := fetchSecrets(f)
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
func fetchSecrets(f types.Factory) (*v1.SecretList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("v1/secrets"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
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
