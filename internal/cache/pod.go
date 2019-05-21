package cache

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func (p *Pod) ListPodsBySelector(sel *metav1.LabelSelector) map[string]*v1.Pod {
	res := map[string]*v1.Pod{}
	if sel == nil {
		return res
	}

	for fqn, po := range p.pods {
		if matchLabels(po.ObjectMeta.Labels, sel.MatchLabels) {
			res[fqn] = po
		}
	}

	return res
}

// GetPod returns a pod via a label query.
func (p *Pod) GetPod(sel map[string]string) *v1.Pod {
	// if len(sel) == 0 {
	// 	return nil
	// }

	res := p.ListPodsBySelector(&metav1.LabelSelector{MatchLabels: sel})
	for _, v := range res {
		return v
	}

	return nil
}

// PodRefs computes all pods external references.
func (p *Pod) PodRefs(refs ObjReferences) {
	for fqn, po := range p.pods {
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

func (p *Pod) namespaceRefs(ns string, refs ObjReferences) {
	set := StringSet{}
	if s, ok := refs["ns"]; ok {
		set = s
	} else {
		refs["ns"] = set
	}
	set.Add(ns)
}

func (p *Pod) containerRefs(pfqn string, co v1.Container, refs ObjReferences) {
	ns, _ := namespaced(pfqn)

	for _, e := range co.Env {
		if e.ValueFrom == nil {
			continue
		}
		p.secretRefs(ns, pfqn, e.ValueFrom.SecretKeyRef, refs)
		p.configMapRefs(ns, pfqn, e.ValueFrom.ConfigMapKeyRef, refs)
	}

	for _, e := range co.EnvFrom {
		switch {
		case e.ConfigMapRef != nil:
			cmRef := e.ConfigMapRef
			efqn := ResFqn(ConfigMapKey, FQN(ns, cmRef.Name))
			if s, ok := refs[efqn]; ok {
				s.Add(AllKeys)
				continue
			}
			refs[efqn] = StringSet{AllKeys: Blank}
		case e.SecretRef != nil:
			secRef := e.SecretRef
			efqn := ResFqn(SecretKey, FQN(ns, secRef.Name))
			if s, ok := refs[efqn]; ok {
				s.Add(AllKeys)
				continue
			}
			refs[efqn] = StringSet{AllKeys: Blank}
		}
	}
}

func (p *Pod) secretRefs(ns, pfqn string, ref *v1.SecretKeySelector, refs ObjReferences) {
	if ref == nil {
		return
	}

	key := ResFqn(SecretKey, FQN(ns, ref.LocalObjectReference.Name))
	if s, ok := refs[key]; ok {
		s.Add(ref.Key)
		return
	}
	refs[key] = StringSet{ref.Key: Blank}
}

func (p *Pod) configMapRefs(ns, pfqn string, ref *v1.ConfigMapKeySelector, refs ObjReferences) {
	if ref == nil {
		return
	}

	key := ResFqn(ConfigMapKey, FQN(ns, ref.LocalObjectReference.Name))
	if s, ok := refs[key]; ok {
		s.Add(ref.Key)
		return
	}
	refs[key] = StringSet{ref.Key: Blank}
}

func (*Pod) volumeRefs(ns string, vv []v1.Volume, refs ObjReferences) {
	for _, v := range vv {
		sv := v.VolumeSource.Secret
		if sv != nil {
			sfqn := FQN(ns, sv.SecretName)
			addKeys(SecretKey, sfqn, sv.Items, refs)
			continue
		}

		cmv := v.VolumeSource.ConfigMap
		if cmv != nil {
			sfqn := FQN(ns, cmv.LocalObjectReference.Name)
			addKeys(ConfigMapKey, sfqn, cmv.Items, refs)
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func addKeys(kind, rfqn string, items []v1.KeyToPath, refs ObjReferences) {
	set := StringSet{}
	key := ResFqn(kind, rfqn)
	if s, ok := refs[key]; ok {
		set = s
	} else {
		refs[key] = set
	}

	if len(items) == 0 {
		set.Add(AllKeys)
		return
	}

	for _, path := range items {
		set.Add(path.Key)
	}
}
