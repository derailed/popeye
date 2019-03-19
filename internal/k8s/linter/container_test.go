package linter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestContainerCheckResources(t *testing.T) {
	uu := []struct {
		request  bool
		limit    bool
		issues   int
		severity Level
	}{
		{request: true, limit: true, issues: 0},
		{request: true, limit: false, issues: 0},
		{request: false, limit: true, issues: 0},
		{request: false, limit: false, issues: 1, severity: InfoLevel},
	}

	for _, u := range uu {
		co := v1.Container{Name: "c1"}
		if u.request {
			co.Resources = v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: resource.Quantity{},
				},
			}
		}
		if u.limit {
			co.Resources = v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU: resource.Quantity{},
				},
			}
		}

		l := NewContainer()
		l.checkResources(co)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}

func TestContainerCheckProbes(t *testing.T) {
	uu := []struct {
		liveness  bool
		readiness bool
		issues    int
		severity  Level
	}{
		{liveness: true, readiness: true, issues: 0},
		{liveness: true, readiness: false, issues: 1, severity: InfoLevel},
		{liveness: false, readiness: true, issues: 1, severity: InfoLevel},
		{liveness: false, readiness: false, issues: 2, severity: InfoLevel},
	}

	for _, u := range uu {
		co := v1.Container{Name: "c1"}
		if u.liveness {
			co.LivenessProbe = &v1.Probe{}
		}
		if u.readiness {
			co.ReadinessProbe = &v1.Probe{}
		}

		l := NewContainer()
		l.checkProbes(co)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}

func TestContainerCheckImageTags(t *testing.T) {
	uu := []struct {
		image    string
		issues   int
		severity Level
	}{
		{image: "cool:1.2.3", issues: 0},
		{image: "fred", issues: 1, severity: WarnLevel},
		{image: "fred:latest", issues: 1, severity: WarnLevel},
	}

	for _, u := range uu {
		co := v1.Container{
			Name:  "c1",
			Image: u.image,
		}

		l := NewContainer()
		l.checkImageTags(co)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}

func TestContainerCheckNamedPorts(t *testing.T) {
	uu := []struct {
		port     string
		issues   int
		severity Level
	}{
		{port: "cool", issues: 0},
		{port: "", issues: 1, severity: InfoLevel},
	}

	for _, u := range uu {
		co := v1.Container{
			Name: "c1",
			Ports: []v1.ContainerPort{
				{Name: u.port},
			},
		}

		l := NewContainer()
		l.checkNamedPorts(co)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}
