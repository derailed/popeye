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

// ListNodes list all included Nodes.
func ListNodes(ctx context.Context) (map[string]*v1.Node, error) {
	return listAllNodes(ctx)
}

// ListAllNodes fetch all Nodes on the cluster.
func listAllNodes(ctx context.Context) (map[string]*v1.Node, error) {
	ll, err := fetchNodes(ctx)
	if err != nil {
		return nil, err
	}
	nos := make(map[string]*v1.Node, len(ll.Items))
	for i := range ll.Items {
		nos[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return nos, nil
}

// FetchNodes retrieves all Nodes on the cluster.
func fetchNodes(ctx context.Context) (*v1.NodeList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("v1/nodes"))
	oo, err := res.List(ctx)
	if err != nil {
		return nil, err
	}
	var ll v1.NodeList
	for _, o := range oo {
		var no v1.Node
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &no)
		if err != nil {
			return nil, errors.New("expecting node resource")
		}
		ll.Items = append(ll.Items, no)
	}

	return &ll, nil
}
