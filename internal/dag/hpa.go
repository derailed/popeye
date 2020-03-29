package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListHorizontalPodAutoscalers list all included HorizontalPodAutoscalers.
func ListHorizontalPodAutoscalers(f types.Factory, cfg *config.Config) (map[string]*autoscalingv1.HorizontalPodAutoscaler, error) {
	hpas, err := listAllHorizontalPodAutoscalers(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*autoscalingv1.HorizontalPodAutoscaler, len(hpas))
	for fqn, hpa := range hpas {
		if includeNS(f.Client(), hpa.Namespace) {
			res[fqn] = hpa
		}
	}

	return res, nil
}

// ListAllHorizontalPodAutoscalers fetch all HorizontalPodAutoscalers on the cluster.
func listAllHorizontalPodAutoscalers(f types.Factory) (map[string]*autoscalingv1.HorizontalPodAutoscaler, error) {
	ll, err := fetchHorizontalPodAutoscalers(f)
	if err != nil {
		return nil, err
	}

	hpas := make(map[string]*autoscalingv1.HorizontalPodAutoscaler, len(ll.Items))
	for i := range ll.Items {
		hpas[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return hpas, nil
}

// FetchHorizontalPodAutoscalers retrieves all HorizontalPodAutoscalers on the cluster.
func fetchHorizontalPodAutoscalers(f types.Factory) (*autoscalingv1.HorizontalPodAutoscalerList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("autoscaling/v1/horizontalpodautoscalers"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll autoscalingv1.HorizontalPodAutoscalerList
	for _, o := range oo {
		var hpa autoscalingv1.HorizontalPodAutoscaler
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &hpa)
		if err != nil {
			return nil, errors.New("expecting hpa resource")
		}
		ll.Items = append(ll.Items, hpa)
	}

	return &ll, nil
}
