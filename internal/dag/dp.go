package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListDeployments list all included Deployments.
func ListDeployments(f types.Factory, cfg *config.Config) (map[string]*appsv1.Deployment, error) {
	dps, err := listAllDeployments(f)
	if err != nil {
		return map[string]*appsv1.Deployment{}, err
	}

	res := make(map[string]*appsv1.Deployment, len(dps))
	for fqn, dp := range dps {
		if includeNS(f.Client(), dp.Namespace) {
			res[fqn] = dp
		}
	}

	return res, nil
}

// ListAllDeployments fetch all Deployments on the cluster.
func listAllDeployments(f types.Factory) (map[string]*appsv1.Deployment, error) {
	ll, err := fetchDeployments(f)
	if err != nil {
		return nil, err
	}

	dps := make(map[string]*appsv1.Deployment, len(ll.Items))
	for i := range ll.Items {
		dps[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return dps, nil
}

// FetchDeployments retrieves all Deployments on the cluster.
func fetchDeployments(f types.Factory) (*appsv1.DeploymentList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("apps/v1/deployments"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll appsv1.DeploymentList
	for _, o := range oo {
		var dp appsv1.Deployment
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &dp)
		if err != nil {
			return nil, errors.New("expecting deployment resource")
		}
		ll.Items = append(ll.Items, dp)
	}

	return &ll, nil
}
