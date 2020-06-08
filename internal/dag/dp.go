package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListDeployments list all included Deployments.
func ListDeployments(ctx context.Context) (map[string]*appsv1.Deployment, error) {
	return listAllDeployments(ctx)
}

// ListAllDeployments fetch all Deployments on the cluster.
func listAllDeployments(ctx context.Context) (map[string]*appsv1.Deployment, error) {
	ll, err := fetchDeployments(ctx)
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
func fetchDeployments(ctx context.Context) (*appsv1.DeploymentList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.AppsV1().Deployments(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("apps/v1/deployments"))
	oo, err := res.List(ctx)
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
