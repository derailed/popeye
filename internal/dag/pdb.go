package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	pv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListPodDisruptionBudgets list all included PodDisruptionBudgets.
func ListPodDisruptionBudgets(ctx context.Context) (map[string]*pv1beta1.PodDisruptionBudget, error) {
	pdbs, err := listAllPodDisruptionBudgets(ctx)
	if err != nil {
		return nil, err
	}

	f := mustExtractFactory(ctx)
	res := make(map[string]*pv1beta1.PodDisruptionBudget, len(pdbs))
	for fqn, pdb := range pdbs {
		if includeNS(f.Client(), pdb.Namespace) {
			res[fqn] = pdb
		}
	}

	return res, nil
}

// ListAllPodDisruptionBudgets fetch all PodDisruptionBudgets on the cluster.
func listAllPodDisruptionBudgets(ctx context.Context) (map[string]*pv1beta1.PodDisruptionBudget, error) {
	ll, err := fetchPodDisruptionBudgets(ctx)
	if err != nil {
		return nil, err
	}

	pdbs := make(map[string]*pv1beta1.PodDisruptionBudget, len(ll.Items))
	for i := range ll.Items {
		pdbs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return pdbs, nil
}

// fetchPodDisruptionBudgets retrieves all PodDisruptionBudgets on the cluster.
func fetchPodDisruptionBudgets(ctx context.Context) (*pv1beta1.PodDisruptionBudgetList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		return f.Client().DialOrDie().PolicyV1beta1().PodDisruptionBudgets(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("policy/v1beta1/poddisruptionbudgets"))
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll pv1beta1.PodDisruptionBudgetList
	for _, o := range oo {
		var pdb pv1beta1.PodDisruptionBudget
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pdb)
		if err != nil {
			return nil, errors.New("expecting pdb resource")
		}
		ll.Items = append(ll.Items, pdb)
	}

	return &ll, nil
}
