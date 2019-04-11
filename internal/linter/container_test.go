package linter

import (
	"testing"

	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestContainerCheckUtilization(t *testing.T) {
	uu := []struct {
		co       v1.Container
		mx       Metrics
		issues   int
		severity Level
	}{
		{makeContainer("c1", true, true), Metrics{100, 4096}, 0, WarnLevel},
		{makeContainer("c1", true, true), Metrics{5000, 10000}, 1, ErrorLevel},
		{makeContainer("c1", true, true), Metrics{5000, 4000000}, 2, ErrorLevel},
		{makeContainer("c1", false, false), Metrics{5000, 4000000}, 0, ErrorLevel},
		{makeContainer("c1", true, false), Metrics{5000, 4000000}, 2, ErrorLevel},
		{makeContainer("c1", false, true), Metrics{5000, 4000000}, 2, ErrorLevel},
	}

	for _, u := range uu {
		mkl := NewMockLoader()
		m.When(mkl.PodCPULimit()).ThenReturn(float64(80))
		m.When(mkl.PodMEMLimit()).ThenReturn(float64(80))

		l := NewContainer(mkl, nil)
		l.checkUtilization(u.co, u.mx)

		assert.Equal(t, u.issues, len(l.Issues()["c1"]))
		if len(l.Issues()["c1"]) != 0 {
			assert.Equal(t, u.severity, l.Issues()["c1"][0].Severity())
		}
		mkl.VerifyWasCalledOnce().PodCPULimit()
		mkl.VerifyWasCalledOnce().PodMEMLimit()
	}
}

func TestContainerCheckResources(t *testing.T) {
	uu := []struct {
		request  bool
		limit    bool
		issues   int
		severity Level
	}{
		{request: true, limit: true, issues: 0},
		{request: true, limit: false, issues: 1, severity: WarnLevel},
		{request: false, limit: true, issues: 0},
		{request: false, limit: false, issues: 1, severity: WarnLevel},
	}

	for _, u := range uu {
		co := makeContainer("c1", u.request, u.limit)
		l := NewContainer(nil, nil)
		l.checkResources(co)

		assert.Equal(t, u.issues, len(l.Issues()["c1"]))
		if len(l.Issues()["c1"]) != 0 {
			assert.Equal(t, u.severity, l.MaxSeverity("c1"))
		}
	}
}

func TestContainerCheckProbes(t *testing.T) {
	uu := []struct {
		liveness  bool
		readiness bool
		namedPort bool
		issues    int
		severity  Level
	}{
		{liveness: true, readiness: true, issues: 0},
		{liveness: true, readiness: false, issues: 1, severity: WarnLevel},
		{liveness: false, readiness: true, issues: 1, severity: WarnLevel},
		{liveness: false, readiness: false, issues: 1, severity: WarnLevel},
		{liveness: true, readiness: true, namedPort: true, issues: 2, severity: InfoLevel},
	}

	for _, u := range uu {
		co := makeContainer("c1", false, false)
		probe := &v1.Probe{}
		if u.namedPort {
			probe.Handler = v1.Handler{HTTPGet: &v1.HTTPGetAction{Port: intstr.Parse("80")}}
		}
		if u.liveness {
			co.LivenessProbe = probe
		}
		if u.readiness {
			co.ReadinessProbe = probe
		}

		l := NewContainer(nil, nil)
		l.checkProbes(co)

		assert.Equal(t, u.issues, len(l.Issues()["c1"]))
		if len(l.Issues()["c1"]) != 0 {
			assert.Equal(t, u.severity, l.Issues()["c1"][0].Severity())
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
		{image: "fred", issues: 1, severity: ErrorLevel},
		{image: "fred:latest", issues: 1, severity: WarnLevel},
	}

	for _, u := range uu {
		co := makeContainer("c1", false, false)
		co.Image = u.image

		l := NewContainer(nil, nil)
		l.checkImageTags(co)

		assert.Equal(t, u.issues, len(l.Issues()["c1"]))
		if len(l.Issues()["c1"]) != 0 {
			assert.Equal(t, u.severity, l.Issues()["c1"][0].Severity())
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
		{port: "", issues: 1, severity: WarnLevel},
	}

	for _, u := range uu {
		co := makeContainer("c1", false, false)
		co.Ports = []v1.ContainerPort{{Name: u.port}}

		l := NewContainer(nil, nil)
		l.checkNamedPorts(co)

		assert.Equal(t, u.issues, len(l.Issues()["c1"]))
		if len(l.Issues()["c1"]) != 0 {
			assert.Equal(t, u.severity, l.Issues()["c1"][0].Severity())
		}
	}
}

func TestContainerLint(t *testing.T) {
	uu := []struct {
		co     v1.Container
		issues int
	}{
		{makeContainer("c1", false, false), 3},
	}

	for _, u := range uu {
		l := NewContainer(nil, nil)
		l.lint(u.co, false)

		assert.Equal(t, u.issues, len(l.Issues()["c1"]))
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeContainer(n string, reqs, limits bool) v1.Container {
	co := v1.Container{Name: n, Resources: v1.ResourceRequirements{}}

	if reqs {
		co.Resources.Requests = makeRes("100m", "1Mi")
	}
	if limits {
		co.Resources.Limits = makeRes("200m", "2Mi")
	}

	return co
}
