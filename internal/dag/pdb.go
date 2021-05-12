package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	polv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListPodDisruptionBudgets list all included PodDisruptionBudgets.
func ListPodDisruptionBudgets(ctx context.Context) (map[string]*polv1beta1.PodDisruptionBudget, error) {
	return listAllPodDisruptionBudgets(ctx)
}

// ListAllPodDisruptionBudgets fetch all PodDisruptionBudgets on the cluster.
func listAllPodDisruptionBudgets(ctx context.Context) (map[string]*polv1beta1.PodDisruptionBudget, error) {
	ll, err := fetchPodDisruptionBudgets(ctx)
	if err != nil {
		return nil, err
	}
	pdbs := make(map[string]*polv1beta1.PodDisruptionBudget, len(ll.Items))
	for i := range ll.Items {
		pdbs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return pdbs, nil
}

// fetchPodDisruptionBudgets retrieves all PodDisruptionBudgets on the cluster.
func fetchPodDisruptionBudgets(ctx context.Context) (*polv1beta1.PodDisruptionBudgetList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.PolicyV1beta1().PodDisruptionBudgets(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("policy/v1beta1/poddisruptionbudgets"))
	oo, err := res.List(ctx)
	if err != nil {
		return nil, err
	}
	var ll polv1beta1.PodDisruptionBudgetList
	for _, o := range oo {
		var pdb polv1beta1.PodDisruptionBudget
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pdb)
		if err != nil {
			return nil, errors.New("expecting pdb resource")
		}
		ll.Items = append(ll.Items, pdb)
	}

	return &ll, nil
}
