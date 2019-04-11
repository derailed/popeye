package linter

import (
	"context"
	"testing"

	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCMLint(t *testing.T) {
	mkc := NewMockLoader()
	m.When(mkc.ActiveNamespace()).ThenReturn("default")
	m.When(mkc.ListConfigMaps()).ThenReturn(map[string]v1.ConfigMap{
		"default/cm1": makeCM("cm1"),
	}, nil)
	m.When(mkc.ListPods()).ThenReturn(map[string]v1.Pod{
		"default/p1": makePodEnv("p1", "cm1", "fred", false),
	}, nil)

	s := NewCM(mkc, nil)
	s.Lint(context.Background())

	assert.Equal(t, 0, len(s.Issues()["default/cm1"]))
	mkc.VerifyWasCalledOnce().ListConfigMaps()
	mkc.VerifyWasCalledOnce().ListPods()
}

func TestCMLintCMS(t *testing.T) {
	uu := []struct {
		pod   v1.Pod
		cm    v1.ConfigMap
		issue int
	}{
		{makePod("p1"), makeCM("cm1"), 1},
		{makePodVolume("p1", "cm1", "fred", false), makeCM("cm1"), 0},
		{makePodVolume("p1", "cm1", "fred", true), makeCM("cm1"), 0},
		{makePodEnvFrom("p1", "cm1", true), makeCM("cm1"), 0},
		{makePodEnvFrom("p1", "cm1", false), makeCM("cm1"), 0},
		{makePodEnv("p1", "cm1", "fred", false), makeCM("cm1"), 0},
		{makePodEnv("p1", "cm1", "fred", true), makeCM("cm1"), 0},
		{makePodEnv("p1", "cm2", "fred", true), makeCM("cm1"), 1},
		{makePodEnv("p1", "cm1", "fred1", true), makeCM("cm1"), 1},
	}

	for _, u := range uu {
		c := NewCM(nil, nil)
		c.lint(
			map[string]v1.ConfigMap{"default/cm1": u.cm},
			map[string]v1.Pod{"default/p1": u.pod},
		)

		assert.Equal(t, 1, len(c.Issues()))
		assert.Equal(t, u.issue, len(c.Issues()["default/cm1"]))
	}
}

func TestCMCheckContainerRefs(t *testing.T) {
	uu := []struct {
		po      v1.Pod
		key     string
		present bool
		e       *Reference
	}{
		{makePod("v1"), "envFrom", false, nil},
		{makePodEnvFrom("p1", "cm1", true), "envFrom", true, &Reference{name: "default/p1:c1"}},
		{makePodEnvFrom("p1", "cm1", false), "envFrom", true, &Reference{name: "default/p1:c1"}},
		{makePodEnv("p1", "cm1", "fred", false), "env", true, &Reference{
			name: "cm1",
			keys: map[string]struct{}{
				"fred": struct{}{},
			},
		}},
		{makePodEnv("p1", "cm1", "fred", true), "env", false, nil},
	}

	for _, u := range uu {
		refs := References{}
		var c *CM
		c.checkContainerRefs("default/p1", u.po.Spec.Containers, refs)

		v, ok := refs["default/cm1"][u.key]
		if u.present {
			assert.True(t, ok)
			assert.Equal(t, u.e, v)
		}
	}
}

func TestCMCheckVolumes(t *testing.T) {
	uu := []struct {
		po      v1.Pod
		present bool
		e       *Reference
	}{
		// Pod with no volumes.
		{
			makePod("p1"), false, nil,
		},
		// Pod with a volume referencing a cm.
		{
			makePodVolume("p1", "cm1", "fred", false),
			true,
			&Reference{
				name: "default/p1:v1",
				keys: map[string]struct{}{"fred": struct{}{}},
			},
		},
		// Pod with a volume referencing an optional cm.
		{
			makePodVolume("p1", "cm1", "fred", true),
			false,
			nil,
		},
	}

	for _, u := range uu {
		refs := References{}
		var c *CM
		c.checkVolumes("default/p1", u.po.Spec.Volumes, refs)

		v, ok := refs["default/cm1"]["volume"]
		if u.present {
			assert.True(t, ok)
			assert.Equal(t, u.e, v)
		}
	}
}

func TestFQNCM(t *testing.T) {
	assert.Equal(t, "default/cm1", fqnCM(makeCM("cm1")))
}

// ----------------------------------------------------------------------------
// Helpers...

func makeCM(n string) v1.ConfigMap {
	return v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Data: map[string]string{
			"fred": "blee",
		},
	}
}

func makePodVolume(n, cm, key string, optional bool) v1.Pod {
	po := makePod(n)
	po.Spec.Volumes = []v1.Volume{
		{
			Name: "v1",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{Name: cm},
					Items: []v1.KeyToPath{
						{Key: key},
					},
					Optional: &optional,
				},
			},
		},
	}
	return po
}

func makePodEnvFrom(n, cm string, optional bool) v1.Pod {
	po := makePod(n)
	po.Spec.Containers = []v1.Container{
		{
			Name: "c1",
			EnvFrom: []v1.EnvFromSource{
				{
					ConfigMapRef: &v1.ConfigMapEnvSource{
						v1.LocalObjectReference{Name: cm},
						&optional,
					},
				},
			},
		},
	}
	return po
}

func makePodEnv(n, cm, key string, optional bool) v1.Pod {
	po := makePod(n)
	po.Spec.Containers = []v1.Container{
		{
			Name: "c1",
			Env: []v1.EnvVar{
				{
					Name: "BLEE",
					ValueFrom: &v1.EnvVarSource{
						ConfigMapKeyRef: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{Name: cm},
							Key:                  key,
							Optional:             &optional,
						},
					},
				},
			},
		},
	}
	return po
}
