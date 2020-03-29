package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	pv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListPodDisruptionBudgets list all included PodDisruptionBudgets.
func ListPodDisruptionBudgets(f types.Factory, cfg *config.Config) (map[string]*pv1beta1.PodDisruptionBudget, error) {
	pdbs, err := listAllPodDisruptionBudgets(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*pv1beta1.PodDisruptionBudget, len(pdbs))
	for fqn, pdb := range pdbs {
		if includeNS(f.Client(), pdb.Namespace) {
			res[fqn] = pdb
		}
	}

	return res, nil
}

// ListAllPodDisruptionBudgets fetch all PodDisruptionBudgets on the cluster.
func listAllPodDisruptionBudgets(f types.Factory) (map[string]*pv1beta1.PodDisruptionBudget, error) {
	ll, err := fetchPodDisruptionBudgets(f)
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
func fetchPodDisruptionBudgets(f types.Factory) (*pv1beta1.PodDisruptionBudgetList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("policy/v1beta1/poddisruptionbudgets"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
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
