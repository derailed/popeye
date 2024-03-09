// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
)

// Cache tracks commonly used resources.
type Cache struct {
	factory types.Factory
	Config  *config.Config
	DB      *db.DB
	Loader  *db.Loader
	cl      *cache.Cluster
}

func NewCache(dba *db.DB, f types.Factory, c *config.Config) *Cache {
	return &Cache{
		DB:      dba,
		factory: f,
		Config:  c,
		Loader:  db.NewLoader(dba),
	}
}

func (c *Cache) cluster(ctx context.Context) (*cache.Cluster, error) {
	if c.cl != nil {
		return c.cl, nil
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	v, err := dag.ListVersion(ctx)
	if err != nil {
		return nil, err
	}
	c.cl = cache.NewCluster(v)

	return c.cl, nil
}

// Scrubers return a collection of linter scrubbers.
func Scrubers() map[internal.R]ScrubFn {
	return map[internal.R]ScrubFn{
		internal.CL:   NewCluster,
		internal.CM:   NewConfigMap,
		internal.NS:   NewNamespace,
		internal.NO:   NewNode,
		internal.PO:   NewPod,
		internal.PV:   NewPersistentVolume,
		internal.PVC:  NewPersistentVolumeClaim,
		internal.SEC:  NewSecret,
		internal.SVC:  NewService,
		internal.SA:   NewServiceAccount,
		internal.DS:   NewDaemonSet,
		internal.DP:   NewDeployment,
		internal.RS:   NewReplicaSet,
		internal.STS:  NewStatefulSet,
		internal.NP:   NewNetworkPolicy,
		internal.ING:  NewIngress,
		internal.CR:   NewClusterRole,
		internal.CRB:  NewClusterRoleBinding,
		internal.RO:   NewRole,
		internal.ROB:  NewRoleBinding,
		internal.PDB:  NewPodDisruptionBudget,
		internal.HPA:  NewHorizontalPodAutoscaler,
		internal.CJOB: NewCronJob,
		internal.JOB:  NewJob,
		internal.GWC:  NewGatewayClass,
		internal.GW:   NewGateway,
		internal.GWR:  NewHTTPRoute,
	}
}
