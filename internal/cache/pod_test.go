package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestPodRefsVolume(t *testing.T) {
	pods := map[string]*v1.Pod{
		"default/p1": makePodVolume("p1", "cm1", "s1", false),
		"default/p2": makePodVolume("p2", "cm2", "s2", true),
		"default/p3": makePodVolume("p3", "cm2", "s2", false),
	}

	p := NewPod(pods)

	refs := ObjReferences{}
	p.PodRefs(refs)

	assert.Equal(t, 2, len(refs["cm:default/cm1"]))
	assert.Equal(t, 2, len(refs["cm:default/cm2"]))
	assert.Equal(t, 1, len(refs["sec:default/s1"]))
	assert.Equal(t, 1, len(refs["sec:default/s2"]))
	assert.Equal(t, 1, len(refs["ns"]))
}

func TestPodRefsEnvFrom(t *testing.T) {
	pods := map[string]*v1.Pod{
		"default/p1": makePodEnvFrom("p1", "r1", false),
		"default/p2": makePodEnvFrom("p2", "r2", true),
		"default/p3": makePodEnvFrom("p3", "r1", false),
	}

	p := NewPod(pods)

	refs := ObjReferences{}
	p.PodRefs(refs)

	assert.Equal(t, 1, len(refs["cm:default/r1"]))
	assert.Equal(t, 1, len(refs["cm:default/r2"]))
	assert.Equal(t, 1, len(refs["sec:default/r1"]))
	assert.Equal(t, 1, len(refs["sec:default/r2"]))
}

func TestPodRefsEnv(t *testing.T) {
	pods := map[string]*v1.Pod{
		"default/p1": makePodEnv("p1", "r1", false),
		"default/p2": makePodEnv("p2", "r2", true),
	}

	p := NewPod(pods)

	refs := ObjReferences{}
	p.PodRefs(refs)

	assert.Equal(t, 2, len(refs["cm:default/r1"]))
	assert.Equal(t, 2, len(refs["cm:default/r2"]))
	assert.Equal(t, 1, len(refs["sec:default/r1"]))
	assert.Equal(t, 1, len(refs["sec:default/r2"]))
}

func TestNamespaced(t *testing.T) {
	uu := []struct {
		s, ens, en string
	}{
		{"fred/blee", "fred", "blee"},
		{"blee", "", "blee"},
	}

	for _, u := range uu {
		ns, n := namespaced(u.s)
		assert.Equal(t, u.ens, ns)
		assert.Equal(t, u.en, n)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makePodVolume(n, cm, sec string, optional bool) *v1.Pod {
	po := makePod(n)
	po.Spec.Volumes = []v1.Volume{
		{
			Name: "v1",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: cm,
					},
					Items: []v1.KeyToPath{
						{Key: "k1"},
						{Key: "k2"},
					},
					Optional: &optional,
				},
			},
		},
		{
			Name: "v2",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: sec,
					Optional:   &optional,
				},
			},
		},
	}

	return po
}

func makePodEnv(n, ref string, optional bool) *v1.Pod {
	po := makePod(n)
	po.Spec.Containers = []v1.Container{
		{
			Name: "c1",
			Env: []v1.EnvVar{
				{
					Name: "e1",
					ValueFrom: &v1.EnvVarSource{
						ConfigMapKeyRef: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: ref,
							},
							Key:      "k1",
							Optional: &optional,
						},
					},
				},
			},
		},
		{
			Name: "c2",
			Env: []v1.EnvVar{
				{
					Name: "e2",
					ValueFrom: &v1.EnvVarSource{
						ConfigMapKeyRef: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: ref,
							},
							Key:      "k2",
							Optional: &optional,
						},
					},
				},
			},
		},
	}
	po.Spec.InitContainers = []v1.Container{
		{
			Name: "ic1",
			Env: []v1.EnvVar{
				{
					Name: "e1",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{Name: ref},
							Key:                  "k2",
							Optional:             &optional,
						},
					},
				},
			},
		},
	}

	return po
}

func makePodEnvFrom(n, cm string, optional bool) *v1.Pod {
	po := makePod(n)
	po.Spec.Containers = []v1.Container{
		{
			Name: "c1",
			EnvFrom: []v1.EnvFromSource{
				{
					ConfigMapRef: &v1.ConfigMapEnvSource{
						LocalObjectReference: v1.LocalObjectReference{Name: cm},
						Optional:             &optional,
					},
				},
			},
		},
	}
	po.Spec.InitContainers = []v1.Container{
		{
			Name: "ic1",
			EnvFrom: []v1.EnvFromSource{
				{
					SecretRef: &v1.SecretEnvSource{
						LocalObjectReference: v1.LocalObjectReference{Name: cm},
						Optional:             &optional,
					},
				},
			},
		},
	}

	return po
}

func makePod(n string) *v1.Pod {
	po := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
	}

	return &po
}

func makeMxPod(name, cpu, mem string) v1beta1.PodMetrics {
	return v1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Containers: []v1beta1.ContainerMetrics{
			{Name: "c1", Usage: makeRes(cpu, mem)},
			{Name: "c2", Usage: makeRes(cpu, mem)},
			{Name: "c3", Usage: makeRes(cpu, mem)},
		},
	}
}

func makeRes(c, m string) v1.ResourceList {
	cpu, _ := resource.ParseQuantity(c)
	mem, _ := resource.ParseQuantity(m)

	return v1.ResourceList{
		v1.ResourceCPU:    cpu,
		v1.ResourceMemory: mem,
	}
}
