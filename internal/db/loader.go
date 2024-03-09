// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type CastFn[T any] func(o runtime.Object) (*T, error)

type Loader struct {
	DB     *DB
	loaded map[types.GVR]struct{}
	mx     sync.RWMutex
}

func NewLoader(db *DB) *Loader {
	l := Loader{
		DB:     db,
		loaded: make(map[types.GVR]struct{}),
	}

	return &l
}

func (l *Loader) isLoaded(gvr types.GVR) bool {
	l.mx.RLock()
	defer l.mx.RUnlock()

	_, ok := l.loaded[gvr]

	return ok
}

func (l *Loader) setLoaded(gvr types.GVR) {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.loaded[gvr] = struct{}{}
}

// LoadResource loads resource and save to db.
func LoadResource[T metav1.ObjectMetaAccessor](ctx context.Context, l *Loader, gvr types.GVR) error {
	if l.isLoaded(gvr) || gvr == types.BlankGVR {
		return nil
	}
	oo, err := loadResource(ctx, gvr)
	if err != nil {
		return err
	}
	if err = Save[T](ctx, l.DB, gvr, oo); err != nil {
		return err
	}
	l.setLoaded(gvr)

	return nil
}

func Cast[T any](o runtime.Object) (T, error) {
	var r T
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &r); err != nil {
		return r, fmt.Errorf("expecting %T resource but got %T: %w", r, o, err)
	}

	return r, nil
}

func Save[T metav1.ObjectMetaAccessor](ctx context.Context, dba *DB, gvr types.GVR, oo []runtime.Object) error {
	txn := dba.Txn(true)
	defer txn.Commit()
	for _, o := range oo {
		var (
			u   T
			err error
		)
		// !!BOZO!! Dud. Can't hydrate cnp/ccnp from unstructured??
		if gvr.R() == "ciliumnetworkpolicies" || gvr.R() == "ciliumclusterwidenetworkpolicies" {
			bb, err := json.Marshal(o.(*unstructured.Unstructured))
			if err != nil {
				return err
			}
			if err = json.Unmarshal(bb, &u); err != nil {
				return err
			}
		} else {
			u, err = Cast[T](o)
			if err != nil {
				return err
			}
		}
		if err := txn.Insert(gvr.String(), u); err != nil {
			return err
		}
	}

	return nil
}

func (l *Loader) LoadPodMX(ctx context.Context) error {
	pmxGVR := internal.Glossary[internal.PMX]
	if l.isLoaded(pmxGVR) {
		return nil
	}

	c := mustExtractFactory(ctx).Client()

	log.Debug().Msg("PRELOAD PMX")
	ll, err := l.fetchPodsMetrics(c)
	if err != nil {
		return err
	}
	txn := l.DB.Txn(true)
	defer txn.Commit()
	for _, l := range ll.Items {
		if err := txn.Insert(pmxGVR.String(), &l); err != nil {
			return err
		}
	}
	l.setLoaded(pmxGVR)

	return nil
}

func (l *Loader) LoadNodeMX(ctx context.Context) error {
	c := mustExtractFactory(ctx).Client()
	if !c.HasMetrics() {
		return nil
	}

	nmxGVR := internal.Glossary[internal.NMX]
	if l.isLoaded(nmxGVR) {
		return nil
	}
	log.Debug().Msg("PRELOAD NMX")
	ll, err := l.fetchNodesMetrics(c)
	if err != nil {
		return err
	}
	txn := l.DB.Txn(true)
	defer txn.Commit()
	for _, l := range ll.Items {
		if err := txn.Insert(nmxGVR.String(), &l); err != nil {
			return err
		}
	}
	l.setLoaded(nmxGVR)

	return nil
}

func (l *Loader) fetchPodsMetrics(c types.Connection) (*mv1beta1.PodMetricsList, error) {
	vc, err := c.MXDial()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), client.CallTimeout)
	defer cancel()

	return vc.MetricsV1beta1().PodMetricses(c.ActiveNamespace()).List(ctx, metav1.ListOptions{})
}

func (l *Loader) fetchNodesMetrics(c types.Connection) (*mv1beta1.NodeMetricsList, error) {
	vc, err := c.MXDial()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.CallTimeout)
	defer cancel()

	return vc.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
}

func loadResource(ctx context.Context, gvr types.GVR) ([]runtime.Object, error) {
	f := mustExtractFactory(ctx)
	if strings.Contains(gvr.String(), "metrics") && !f.Client().HasMetrics() {
		return nil, nil
	}
	var res dao.Generic
	res.Init(f, gvr)

	return res.List(ctx)
}

func (l *Loader) LoadGeneric(ctx context.Context, gvr types.GVR) error {
	if l.isLoaded(gvr) {
		return nil
	}

	oo, err := l.fetchGeneric(ctx, gvr)
	if err != nil {
		return err
	}
	txn := l.DB.Txn(true)
	defer txn.Commit()
	for _, o := range oo {
		if err := txn.Insert(gvr.String(), o); err != nil {
			return err
		}
	}
	l.setLoaded(gvr)

	return nil
}

func (l *Loader) fetchGeneric(ctx context.Context, gvr types.GVR) ([]runtime.Object, error) {
	f := mustExtractFactory(ctx)
	var res dao.Resource
	res.Init(f, types.NewGVR(gvr.String()))

	return res.List(ctx)
}

// Helpers...

func mustExtractFactory(ctx context.Context) types.Factory {
	f, ok := ctx.Value(internal.KeyFactory).(types.Factory)
	if !ok {
		panic("expecting factory in context")
	}

	return f
}
