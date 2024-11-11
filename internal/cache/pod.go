// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"errors"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	v1 "k8s.io/api/core/v1"
)

// Pod represents a Pod cache.
type Pod struct {
	db *db.DB
}

// NewPod returns a Pod cache.
func NewPod(dba *db.DB) *Pod {
	return &Pod{db: dba}
}

// PodRefs computes all pods external references.
func (p *Pod) PodRefs(refs *sync.Map) error {
	txn, it := p.db.MustITFor(internal.Glossary[internal.PO])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		po, ok := o.(*v1.Pod)
		if !ok {
			return errors.New("expected a pod")
		}
		fqn := client.FQN(po.Namespace, po.Name)
		p.imagePullSecRefs(po.Namespace, po.Spec.ImagePullSecrets, refs)
		namespaceRefs(po.Namespace, refs)
		for _, co := range po.Spec.InitContainers {
			p.containerRefs(fqn, co, refs)
		}
		for _, co := range po.Spec.Containers {
			p.containerRefs(fqn, co, refs)
		}
		p.volumeRefs(po.Namespace, po.Spec.Volumes, refs)
	}

	return nil
}

func (p *Pod) containerRefs(pfqn string, co v1.Container, refs *sync.Map) {
	ns, _ := namespaced(pfqn)

	for _, e := range co.Env {
		if e.ValueFrom == nil {
			continue
		}
		p.secretRefs(ns, e.ValueFrom.SecretKeyRef, refs)
		p.configMapRefs(ns, e.ValueFrom.ConfigMapKeyRef, refs)
	}

	for _, e := range co.EnvFrom {
		switch {
		case e.ConfigMapRef != nil:
			refs.Store(ResFqn(ConfigMapKey, FQN(ns, e.ConfigMapRef.Name)), internal.AllKeys)
		case e.SecretRef != nil:
			refs.Store(ResFqn(SecretKey, FQN(ns, e.SecretRef.Name)), internal.AllKeys)
		}
	}
}

func (*Pod) volumeRefs(ns string, vv []v1.Volume, refs *sync.Map) {
	for _, v := range vv {
		sv := v.VolumeSource.Secret
		if sv != nil {
			addKeys(SecretKey, FQN(ns, sv.SecretName), sv.Items, refs)
			continue
		}

		cmv := v.VolumeSource.ConfigMap
		if cmv != nil {
			addKeys(ConfigMapKey, FQN(ns, cmv.LocalObjectReference.Name), cmv.Items, refs)
		}

		if v.VolumeSource.Projected != nil {
			vp := v.VolumeSource.Projected.Sources
			for _, projection := range vp {
				ps := projection.Secret
				if ps != nil {
					addKeys(SecretKey, FQN(ns, ps.Name), ps.Items, refs)
				}

				pcm := projection.ConfigMap
				if pcm != nil {
					addKeys(ConfigMapKey, FQN(ns, pcm.Name), pcm.Items, refs)
				}
			}
		}
	}
}

func (p *Pod) imagePullSecRefs(ns string, sRefs []v1.LocalObjectReference, refs *sync.Map) {
	for _, s := range sRefs {
		key := ResFqn(SecretKey, FQN(ns, s.Name))
		refs.Store(key, internal.AllKeys)
	}
}

func namespaceRefs(ns string, refs *sync.Map) {
	if set, ok := refs.LoadOrStore("ns", internal.StringSet{ns: internal.Blank}); ok {
		set.(internal.StringSet).Add(ns)
	}
}

func (p *Pod) secretRefs(ns string, ref *v1.SecretKeySelector, refs *sync.Map) {
	if ref == nil {
		return
	}
	key := ResFqn(SecretKey, FQN(ns, ref.LocalObjectReference.Name))
	if s, ok := refs.LoadOrStore(key, internal.StringSet{ref.Key: internal.Blank}); ok {
		if ss, ok := s.(internal.StringSet); ok {
			cp := ss.Clone()
			cp.Add(ref.Key)
			refs.Store(key, cp)
		}
	}
}

func (p *Pod) configMapRefs(ns string, ref *v1.ConfigMapKeySelector, refs *sync.Map) {
	if ref == nil {
		return
	}
	key := ResFqn(ConfigMapKey, FQN(ns, ref.LocalObjectReference.Name))
	if s, ok := refs.LoadOrStore(key, internal.StringSet{ref.Key: internal.Blank}); ok {
		if ss, ok := s.(internal.StringSet); ok {
			cp := ss.Clone()
			cp.Add(ref.Key)
			refs.Store(key, cp)
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func addKeys(kind, rfqn string, items []v1.KeyToPath, refs *sync.Map) {
	set := make(internal.StringSet, len(items))
	if len(items) == 0 {
		set.Add(internal.All)
	}
	for _, path := range items {
		set.Add(path.Key)
	}

	key := ResFqn(kind, rfqn)
	if s, ok := refs.LoadOrStore(ResFqn(kind, rfqn), set); ok {
		if ss, ok := s.(internal.StringSet); ok {
			cp := ss.Clone()
			cp.AddAll(set)
			refs.Store(key, cp)
		}
	}
}
