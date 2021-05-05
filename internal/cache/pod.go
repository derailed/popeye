package cache

import (
	"sync"

	"github.com/derailed/popeye/internal"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// Pod represents a Pod cache.
type Pod struct {
	pods map[string]*v1.Pod
}

// NewPod returns a Pod cache.
func NewPod(pods map[string]*v1.Pod) *Pod {
	return &Pod{pods: pods}
}

// ListPods return available pods.
func (p *Pod) ListPods() map[string]*v1.Pod {
	return p.pods
}

// ListPodsBySelector list all pods matching the given selector.
func (p *Pod) ListPodsBySelector(ns string, sel *metav1.LabelSelector) map[string]*v1.Pod {
	res := map[string]*v1.Pod{}
	if sel == nil || sel.Size() == 0 {
		return res
	}
	for fqn, po := range p.pods {
		if po.Namespace != ns {
			continue
		}
		if s, err := metav1.LabelSelectorAsSelector(sel); err == nil && s.Matches(labels.Set(po.Labels)) {
			res[fqn] = po
		}
	}

	return res
}

// GetPod returns a pod via a label query.
func (p *Pod) GetPod(ns string, sel map[string]string) *v1.Pod {
	res := p.ListPodsBySelector(ns, &metav1.LabelSelector{MatchLabels: sel})
	for _, v := range res {
		return v
	}

	return nil
}

// PodRefs computes all pods external references.
func (p *Pod) PodRefs(refs *sync.Map) {
	for fqn, po := range p.pods {
		p.imagePullSecRefs(po.Namespace, po.Spec.ImagePullSecrets, refs)
		p.namespaceRefs(po.Namespace, refs)
		for _, co := range po.Spec.InitContainers {
			p.containerRefs(fqn, co, refs)
		}
		for _, co := range po.Spec.Containers {
			p.containerRefs(fqn, co, refs)
		}
		p.volumeRefs(po.Namespace, po.Spec.Volumes, refs)
	}
}

func (p *Pod) imagePullSecRefs(ns string, sRefs []v1.LocalObjectReference, refs *sync.Map) {
	for _, s := range sRefs {
		key := ResFqn(SecretKey, FQN(ns, s.Name))
		refs.Store(key, internal.AllKeys)
	}
}

func (p *Pod) namespaceRefs(ns string, refs *sync.Map) {
	if set, ok := refs.LoadOrStore("ns", internal.StringSet{ns: internal.Blank}); ok {
		set.(internal.StringSet).Add(ns)
	}
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
