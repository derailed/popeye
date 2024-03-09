// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cilium"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/db/schema"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	"github.com/hashicorp/go-memdb"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

func NewTestDB() (*db.DB, error) {
	initLinters()
	initCiliumLinters()
	d, err := memdb.NewMemDB(schema.Init())
	if err != nil {
		return nil, err
	}

	return db.NewDB(d), nil
}

func initCiliumLinters() {
	internal.Glossary[cilium.CID] = types.NewGVR("cilium.io/v2/ciliumidentities")
	internal.Glossary[cilium.CEP] = types.NewGVR("cilium.io/v2/ciliumendpoints")
	internal.Glossary[cilium.CNP] = types.NewGVR("cilium.io/v2/ciliumnetworkpolicies")
	internal.Glossary[cilium.CCNP] = types.NewGVR("cilium.io/v2/ciliumclusterwidenetworkpolicies")
}

func initLinters() {
	internal.Glossary = internal.Linters{
		internal.CM:   types.NewGVR("v1/configmaps"),
		internal.EP:   types.NewGVR("v1/endpoints"),
		internal.NS:   types.NewGVR("v1/namespaces"),
		internal.NO:   types.NewGVR("v1/nodes"),
		internal.PV:   types.NewGVR("v1/persistentvolumes"),
		internal.PVC:  types.NewGVR("v1/persistentvolumeclaims"),
		internal.PO:   types.NewGVR("v1/pods"),
		internal.SEC:  types.NewGVR("v1/secrets"),
		internal.SA:   types.NewGVR("v1/serviceaccounts"),
		internal.SVC:  types.NewGVR("v1/services"),
		internal.DS:   types.NewGVR("apps/v1/daemonsets"),
		internal.DP:   types.NewGVR("apps/v1/deployments"),
		internal.RS:   types.NewGVR("apps/v1/replicasets"),
		internal.STS:  types.NewGVR("apps/v1/statefulsets"),
		internal.CR:   types.NewGVR("rbac.authorization.k8s.io/v1/clusterroles"),
		internal.CRB:  types.NewGVR("rbac.authorization.k8s.io/v1/clusterrolebindings"),
		internal.RO:   types.NewGVR("rbac.authorization.k8s.io/v1/roles"),
		internal.ROB:  types.NewGVR("rbac.authorization.k8s.io/v1/rolebindings"),
		internal.ING:  types.NewGVR("networking.k8s.io/v1/ingresses"),
		internal.NP:   types.NewGVR("networking.k8s.io/v1/networkpolicies"),
		internal.PDB:  types.NewGVR("policy/v1/poddisruptionbudgets"),
		internal.HPA:  types.NewGVR("autoscaling/v1/horizontalpodautoscalers"),
		internal.PMX:  types.NewGVR("metrics.k8s.io/v1beta1/podmetrics"),
		internal.NMX:  types.NewGVR("metrics.k8s.io/v1beta1/nodemetrics"),
		internal.CJOB: types.NewGVR("batch/v1/cronjobs"),
		internal.JOB:  types.NewGVR("batch/v1/jobs"),
		internal.GW:   types.NewGVR("gateway.networking.k8s.io/v1/gateways"),
		internal.GWC:  types.NewGVR("gateway.networking.k8s.io/v1/gatewayclasses"),
		internal.GWR:  types.NewGVR("gateway.networking.k8s.io/v1/httproutes"),
	}
}

func MakeRes(c, m string) v1.ResourceList {
	return v1.ResourceList{
		v1.ResourceCPU:    *MakeQty(c),
		v1.ResourceMemory: *MakeQty(m),
	}
}

func MakeQty(s string) *resource.Quantity {
	if s == "" {
		return nil
	}

	qty, _ := resource.ParseQuantity(s)
	return &qty
}

func ToQty(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)

	return q
}

func LoadRes[T any](p string) (*T, error) {
	bb, err := os.ReadFile(filepath.Join("testdata", p))
	if err != nil {
		return nil, err
	}
	var l T
	if err := yaml.Unmarshal(bb, &l); err != nil {
		return nil, err
	}

	return &l, nil
}

func LoadDB[T metav1.ObjectMetaAccessor](ctx context.Context, dba *db.DB, p string, gvr types.GVR) error {
	ucc, err := LoadRes[unstructured.UnstructuredList](p)
	if err != nil {
		return err
	}
	cc := make([]runtime.Object, 0, len(ucc.Items))
	for i := range ucc.Items {
		u := ucc.Items[i]
		cc = append(cc, &u)
	}

	return db.Save[T](ctx, dba, gvr, cc)
}

func MakeCollector(t *testing.T) *issues.Collector {
	return issues.NewCollector(loadCodes(t), MakeConfig(t))
}

func MakeCtx(t *testing.T) context.Context {
	return context.WithValue(context.Background(), internal.KeyConfig, MakeConfig(t))
}

func loadCodes(t *testing.T) *issues.Codes {
	codes, err := issues.LoadCodes()
	assert.Nil(t, err)

	return codes
}

func MakeConfig(t *testing.T) *config.Config {
	c, err := config.NewConfig(config.NewFlags())
	assert.Nil(t, err)
	return c
}

func MakeContext(gvr, section string) context.Context {
	return context.WithValue(context.Background(), internal.KeyRunInfo, internal.RunInfo{
		Section:    section,
		SectionGVR: types.NewGVR(gvr),
	})
}
