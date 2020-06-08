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

// ListPods list all filtered pods.
func ListPods(ctx context.Context) (map[string]*v1.Pod, error) {
	return listAllPods(ctx)
}

// ListAllPods fetch all Pods on the cluster.
func listAllPods(ctx context.Context) (map[string]*v1.Pod, error) {
	ll, err := fetchPods(ctx)
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
func fetchPods(ctx context.Context) (*v1.PodList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.CoreV1().Pods(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("v1/pods"))
	oo, err := res.List(ctx)
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
