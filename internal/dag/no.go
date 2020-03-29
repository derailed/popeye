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

// ListNodes list all included Nodes.
func ListNodes(f types.Factory, cfg *config.Config) (map[string]*v1.Node, error) {
	nos, err := listAllNodes(f)
	if err != nil {
		return nil, err
	}
	res := make(map[string]*v1.Node, len(nos))
	for fqn, no := range nos {
		res[fqn] = no
	}

	return res, nil
}

// ListAllNodes fetch all Nodes on the cluster.
func listAllNodes(f types.Factory) (map[string]*v1.Node, error) {
	ll, err := fetchNodes(f)
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
func fetchNodes(f types.Factory) (*v1.NodeList, error) {
	// return c.DialOrDie().CoreV1().Nodes().List(metav1.ListOptions{})
	var res dao.Resource
	res.Init(f, client.NewGVR("v1/nodes"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
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
