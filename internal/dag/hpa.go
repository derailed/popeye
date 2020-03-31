package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListHorizontalPodAutoscalers list all included HorizontalPodAutoscalers.
func ListHorizontalPodAutoscalers(ctx context.Context) (map[string]*autoscalingv1.HorizontalPodAutoscaler, error) {
	hpas, err := listAllHorizontalPodAutoscalers(ctx)
	if err != nil {
		return nil, err
	}

	f := mustExtractFactory(ctx)
	res := make(map[string]*autoscalingv1.HorizontalPodAutoscaler, len(hpas))
	for fqn, hpa := range hpas {
		if includeNS(f.Client(), hpa.Namespace) {
			res[fqn] = hpa
		}
	}

	return res, nil
}

// ListAllHorizontalPodAutoscalers fetch all HorizontalPodAutoscalers on the cluster.
func listAllHorizontalPodAutoscalers(ctx context.Context) (map[string]*autoscalingv1.HorizontalPodAutoscaler, error) {
	ll, err := fetchHorizontalPodAutoscalers(ctx)
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
func fetchHorizontalPodAutoscalers(ctx context.Context) (*autoscalingv1.HorizontalPodAutoscalerList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		return f.Client().DialOrDie().AutoscalingV1().HorizontalPodAutoscalers(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("autoscaling/v1/horizontalpodautoscalers"))
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
