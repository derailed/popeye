package linter

import (
	"context"
	"testing"

	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSecLint(t *testing.T) {
	mkl := NewMockLoader()
	m.When(mkl.ActiveNamespace()).ThenReturn("default")
	m.When(mkl.ListSecrets()).ThenReturn(map[string]v1.Secret{
		"default/s1": makeSec("s1"),
	}, nil)
	m.When(mkl.ListServiceAccounts()).ThenReturn(map[string]v1.ServiceAccount{
		"default/sa1": makeSA("sa1"),
	}, nil)
	m.When(mkl.ListPods()).ThenReturn(map[string]v1.Pod{
		"default/p1": makePodSecEnv("p1", "s1", "fred", false),
	}, nil)

	s := NewSecret(mkl, nil)
	s.Lint(context.Background())

	assert.Equal(t, 0, len(s.Issues()["default/s1"]))

	mkl.VerifyWasCalledOnce().ListSecrets()
	mkl.VerifyWasCalledOnce().ListServiceAccounts()
	mkl.VerifyWasCalledOnce().ListPods()
}

func TestSecLintSecrets(t *testing.T) {
	uu := []struct {
		sec   v1.Secret
		pod   v1.Pod
		sa    v1.ServiceAccount
		issue int
	}{
		{makeSec("s1"), makePod("p1"), makeSA("sa1"), 1},
		{makeSec("s1"), makePod("p1"), makeSASec("sa1", "s1"), 0},
		{makeSec("s1"), makePodPullSec("p1", "s1"), makeSASec("sa1", "s1"), 0},
		{makeSec("s1"), makePodPullSec("p1", "s2"), makeSASec("sa1", "s2"), 1},
		{makeSec("s1"), makePodPullSec("p1", "s1"), makeSASec("sa1", "s3"), 1},
		{makeSec("s1"), makePodPullSec("p1", "s3"), makeSASec("sa1", "s1"), 0},
		{makeSec("s1"), makePodSecVol("p1", "s1", "blee", true), makeSASec("sa1", "s2"), 1},
		{makeSec("s1"), makePodSecVol("p1", "s1", "blee", true), makeSASec("sa1", "s1"), 0},
		{makeSec("s1"), makePodSecVol("p1", "s2", "blee", true), makeSAPull("sa1", "s1"), 0},
	}

	for _, u := range uu {
		s := NewSecret(nil, nil)
		s.lint(
			map[string]v1.Secret{"default/s1": u.sec},
			map[string]v1.Pod{"default/p1": u.pod},
			map[string]v1.ServiceAccount{"default/sa1": u.sa},
		)

		assert.Equal(t, 1, len(s.Issues()))
		assert.Equal(t, u.issue, len(s.Issues()["default/s1"]))
	}
}

func TestPullImageSecrets(t *testing.T) {
	uu := []struct {
		po      v1.Pod
		key     string
		present bool
		e       *Reference
	}{
		{makePodPullSec("v1", "s1"), "pull", false, nil},
	}

	for _, u := range uu {
		refs := References{}
		var s *Secret
		s.checkPullImageSecrets(u.po, refs)

		v, ok := refs["default/s1"][u.key]
		if u.present {
			assert.True(t, ok)
			assert.Equal(t, u.e, v)
		}
	}
}

func TestSecCheckContainerRefs(t *testing.T) {
	uu := []struct {
		po      v1.Pod
		key     string
		present bool
		e       *Reference
	}{
		{makePod("v1"), "envFrom", false, nil},
		{makePodSecEnv("p1", "s1", "fred", false), "env", true, &Reference{
			name: "s1",
			keys: map[string]struct{}{
				"fred": {},
			},
		}},
		{makePodEnv("p1", "s1", "fred", true), "env", false, nil},
	}

	for _, u := range uu {
		refs := References{}
		var s *Secret
		s.checkContainerRefs(podFQN(u.po), u.po.Spec.Containers, refs)

		v, ok := refs["default/s1"][u.key]
		if u.present {
			assert.True(t, ok)
			assert.Equal(t, u.e, v)
		}
	}
}

func TestSecCheckVolumes(t *testing.T) {
	uu := []struct {
		po      v1.Pod
		key     string
		present bool
		e       *Reference
	}{
		{
			makePod("p1"), "volume", false, nil,
		},
		{
			makePodSecVol("p1", "s1", "fred", false), "volume", true, &Reference{
				name: "default/p1:v1",
				keys: map[string]struct{}{"fred": {}},
			},
		},
		{
			makePodSecVol("p1", "s1", "fred", true), "volume", false, nil,
		},
	}

	for _, u := range uu {
		refs := References{}
		var s *Secret
		s.checkVolumes(podFQN(u.po), u.po.Spec.Volumes, refs)

		v, ok := refs["default/s1"][u.key]
		if u.present {
			assert.True(t, ok)
			assert.Equal(t, u.e, v)
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeSA(n string) v1.ServiceAccount {
	return v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
	}
}

func makeSASec(n, s string) v1.ServiceAccount {
	sa := makeSA(n)
	sa.Secrets = []v1.ObjectReference{{Namespace: "default", Name: s}}

	return sa
}

func makeSAPull(n, s string) v1.ServiceAccount {
	sa := makeSA(n)
	sa.ImagePullSecrets = []v1.LocalObjectReference{{Name: s}}

	return sa
}

func makeSec(n string) v1.Secret {
	return v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Data: map[string][]byte{
			"fred": []byte("blee"),
		},
	}
}

func makePodPullSec(n, sec string) v1.Pod {
	po := makePod(n)
	po.Spec.ImagePullSecrets = []v1.LocalObjectReference{
		{
			Name: sec,
		},
	}

	return po
}

func makePodSecVol(n, sec, key string, optional bool) v1.Pod {
	po := makePod(n)
	po.Spec.Volumes = []v1.Volume{
		{
			Name: "v1",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: sec,
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

func makePodSecEnv(n, sec, key string, optional bool) v1.Pod {
	po := makePod(n)
	po.Spec.Containers = []v1.Container{
		{
			Name: "c1",
			Env: []v1.EnvVar{
				{
					Name: "BLEE",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{Name: sec},
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
