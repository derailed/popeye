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

// ListPods list all filtered pods.
func ListPods(f types.Factory, cfg *config.Config) (map[string]*v1.Pod, error) {
	pods, err := listAllPods(f)
	if err != nil {
		return map[string]*v1.Pod{}, err
	}
	res := make(map[string]*v1.Pod, len(pods))
	for fqn, po := range pods {
		if includeNS(f.Client(), po.Namespace) {
			res[fqn] = po
		}
	}

	return res, nil
}

// ListAllPods fetch all Pods on the cluster.
func listAllPods(f types.Factory) (map[string]*v1.Pod, error) {
	ll, err := fetchPods(f)
	if err != nil {
		return nil, err
	}
	pods := make(map[string]*v1.Pod, len(ll.Items))
	for i := range ll.Items {
		pods[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return pods, nil
}

// FetchPods retrieves all Pods on the cluster.
func fetchPods(f types.Factory) (*v1.PodList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("v1/pods"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll v1.PodList
	for _, o := range oo {
		var po v1.Pod
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &po)
		if err != nil {
			return nil, errors.New("expecting pod resource")
		}
		ll.Items = append(ll.Items, po)
	}

	return &ll, nil
}
