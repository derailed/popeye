package sanitize

import (
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestContainerCheckUtilization(t *testing.T) {
	uu := map[string]struct {
		co     v1.Container
		mx     client.Metrics
		issues int
	}{
		"cool": {
			co: makeContainer("c1", coOpts{
				rcpu: "10m",
				rmem: "10Mi",
				lcpu: "10m",
				lmem: "10Mi",
			}),
			mx: client.Metrics{CurrentCPU: toQty("1m"), CurrentMEM: toQty("1Mi")},
		},
		"cpuOver": {
			co: makeContainer("c1", coOpts{
				rcpu: "50m",
				rmem: "10Mi",
				lcpu: "100m",
				lmem: "10Mi",
			}),
			mx:     client.Metrics{CurrentCPU: toQty("200m"), CurrentMEM: toQty("2Mi")},
			issues: 1,
		},
		"memOver": {
			co: makeContainer("c1", coOpts{
				rcpu: "50m",
				rmem: "10Mi",
				lcpu: "100m",
				lmem: "10Mi",
			}),
			mx:     client.Metrics{CurrentCPU: toQty("10m"), CurrentMEM: toQty("20Mi")},
			issues: 1,
		},
		"bothOver": {
			co: makeContainer("c1", coOpts{
				rcpu: "100m",
				rmem: "10Mi",
				lcpu: "100m",
				lmem: "10Mi",
			}),
			mx:     client.Metrics{CurrentCPU: toQty("5"), CurrentMEM: toQty("20Mi")},
			issues: 2,
		},
		"LimOver": {
			co: makeContainer("c1", coOpts{
				rcpu: "",
				rmem: "",
				lcpu: "100m",
				lmem: "10Mi",
			}),
			mx:     client.Metrics{CurrentCPU: toQty("5"), CurrentMEM: toQty("20Mi")},
			issues: 2,
		},
	}

	ctx := makeContext("containers", "container")
	ctx = internal.WithFQN(ctx, "default/p1")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := NewContainer("default/p1", newRangeCollector(t))
			ctx = internal.WithGroup(ctx, client.NewGVR("containers"), u.co.Name)
			c.checkUtilization(ctx, u.co, u.mx)

			assert.Equal(t, u.issues, len(c.Outcome().For("default/p1", "c1")))
		})
	}
}

func TestContainerCheckResources(t *testing.T) {
	uu := map[string]struct {
		request  bool
		limit    bool
		issues   int
		severity config.Level
	}{
		"cool":  {request: true, limit: true, issues: 0},
		"noLim": {request: true, issues: 1, severity: config.WarnLevel},
		"noReq": {limit: true, issues: 0},
		"none":  {issues: 1, severity: config.WarnLevel},
	}

	ctx := makeContext("containers", "container")
	for k := range uu {
		u := uu[k]
		opts := coOpts{}
		if u.request {
			opts.rcpu = "100m"
			opts.rmem = "10Mi"
		}
		if u.limit {
			opts.lcpu = "100m"
			opts.lmem = "10Mi"
		}
		co := makeContainer("c1", opts)
		l := NewContainer("default/p1", newRangeCollector(t))

		t.Run(k, func(t *testing.T) {
			ctx = internal.WithFQN(ctx, "default/p1")
			ctx = internal.WithGroup(ctx, client.NewGVR("containers"), co.Name)
			l.checkResources(ctx, co)

			assert.Equal(t, u.issues, len(l.Outcome()["default/p1"]))
			if len(l.Outcome()["default/p1"]) != 0 {
				assert.Equal(t, u.severity, l.Outcome()["default/p1"].MaxSeverity())
			}
		})
	}
}

func TestContainerCheckProbes(t *testing.T) {
	uu := map[string]struct {
		liveness  bool
		readiness bool
		namedPort bool
		issues    int
		severity  config.Level
	}{
		"cool":       {liveness: true, readiness: true},
		"noReady":    {liveness: true, issues: 1, severity: config.WarnLevel},
		"noLive":     {readiness: true, issues: 1, severity: config.WarnLevel},
		"noneProbes": {issues: 1, severity: config.WarnLevel},
		"Unnamed":    {liveness: true, readiness: true, namedPort: true, issues: 2, severity: config.InfoLevel},
	}

	ctx := makeContext("containers", "container")
	for k := range uu {
		u := uu[k]
		co := makeContainer("c1", coOpts{})
		probe := &v1.Probe{}
		if u.namedPort {
			probe.ProbeHandler = v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{Port: intstr.Parse("80")}}
		}
		if u.liveness {
			co.LivenessProbe = probe
		}
		if u.readiness {
			co.ReadinessProbe = probe
		}

		c := NewContainer("default/p1", newRangeCollector(t))
		t.Run(k, func(t *testing.T) {
			c.checkProbes(ctx, co)

			if len(c.Outcome()["default/p1"]) != 0 {
				assert.Equal(t, u.issues, len(c.Outcome().For("default/p1", "c1")))
				assert.Equal(t, u.severity, c.Outcome().For("default/p1", "c1").MaxSeverity())
			}
		})
	}
}

func TestContainerCheckImageTags(t *testing.T) {
	uu := map[string]struct {
		image    string
		pissues  int
		issues   int
		severity config.Level
	}{
		"cool":   {image: "cool:1.2.3", issues: 0},
		"noRev":  {pissues: 1, image: "fred", issues: 1, severity: config.ErrorLevel},
		"latest": {pissues: 1, image: "fred:latest", issues: 1, severity: config.WarnLevel},
	}

	ctx := makeContext("containers", "container")
	ctx = internal.WithFQN(ctx, "default/p1")
	ctx = internal.WithGroup(ctx, client.NewGVR("containers"), "c1")
	for k := range uu {
		u := uu[k]
		co := makeContainer("c1", coOpts{})
		co.Image = u.image

		l := NewContainer("default/p1", newRangeCollector(t))
		t.Run(k, func(t *testing.T) {
			l.checkImageTags(ctx, co.Image)

			assert.Equal(t, u.pissues, len(l.Outcome()["default/p1"]))
			if len(l.Outcome()["default/p1"]) != 0 {
				assert.Equal(t, u.issues, len(l.Outcome().For("default/p1", "c1")))
				assert.Equal(t, u.severity, l.Outcome().For("default/p1", "c1").MaxSeverity())
			}
		})
	}
}

func TestContainerCheckImageRegistry(t *testing.T) {
	uu := map[string]struct {
		image    string
		pissues  int
		issues   int
		severity config.Level
	}{
		"dockerDefault": {image: "dockerhub:1.2.3", issues: 0},
		"cool":          {image: "docker.io/cool:1.2.3", issues: 0},
		"wrongRegistry": {pissues: 1, image: "wrong-registry.io/fred", issues: 1, severity: config.ErrorLevel},
	}

	ctx := makeContext("containers", "container")
	ctx = internal.WithFQN(ctx, "default/p1")
	ctx = internal.WithGroup(ctx, client.NewGVR("containers"), "c1")
	for k := range uu {
		u := uu[k]
		co := makeContainer("c1", coOpts{})
		co.Image = u.image

		l := NewContainer("default/p1", newRangeCollectorWithRegistry(t))
		t.Run(k, func(t *testing.T) {
			l.checkImageTags(ctx, co.Image)

			assert.Equal(t, u.pissues, len(l.Outcome()["default/p1"]))
			if len(l.Outcome()["default/p1"]) != 0 {
				assert.Equal(t, u.issues, len(l.Outcome().For("default/p1", "c1")))
				assert.Equal(t, u.severity, l.Outcome().For("default/p1", "c1").MaxSeverity())
			}
		})
	}
}

func TestContainerCheckNamedPorts(t *testing.T) {
	uu := map[string]struct {
		port     string
		issues   int
		severity config.Level
	}{
		"named":  {port: "cool", issues: 0},
		"unamed": {port: "", issues: 1, severity: config.WarnLevel},
	}

	ctx := makeContext("containers", "container")
	ctx = internal.WithFQN(ctx, "p1")
	ctx = internal.WithGroup(ctx, client.NewGVR("v1/pods"), "p1")
	for k := range uu {
		u := uu[k]
		co := makeContainer("c1", coOpts{})
		co.Ports = []v1.ContainerPort{{Name: u.port}}

		l := NewContainer("p1", newRangeCollector(t))
		t.Run(k, func(t *testing.T) {
			l.checkNamedPorts(ctx, co)

			assert.Equal(t, u.issues, len(l.Outcome()["p1"]))
			if len(l.Outcome()["c1"]) != 0 {
				assert.Equal(t, u.severity, l.Outcome()["c1"].MaxSeverity())
			}
		})
	}
}

func TestContainerSanitize(t *testing.T) {
	uu := map[string]struct {
		co     v1.Container
		issues int
	}{
		"NoImgNoProbs": {makeContainer("c1", coOpts{}), 3},
	}

	ctx := makeContext("containers", "container")
	for k := range uu {
		u := uu[k]
		c := NewContainer("default/p1", newRangeCollector(t))
		t.Run(k, func(t *testing.T) {
			c.sanitize(ctx, u.co, true)

			assert.Equal(t, 3, len(c.Outcome()["default/p1"]))
			assert.Equal(t, u.issues, len(c.Outcome().For("default/p1", "c1")))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type rangeCollector struct {
	*issues.Collector
}

func newRangeCollector(t *testing.T) *rangeCollector {
	return &rangeCollector{issues.NewCollector(loadCodes(t), makeConfig(t))}
}

func newRangeCollectorWithRegistry(t *testing.T) *rangeCollector {
	cfg := makeConfig(t)
	cfg.Registries = append(cfg.Registries, "docker.io")
	return &rangeCollector{issues.NewCollector(loadCodes(t), cfg)}
}

func (*rangeCollector) RestartsLimit() int {
	return 10
}

func (*rangeCollector) PodCPULimit() float64 {
	return 100
}

func (*rangeCollector) PodMEMLimit() float64 {
	return 100
}

type coOpts struct {
	image        string
	rcpu, rmem   string
	lcpu, lmem   string
	lprob, rprob bool
}

func makeContainer(n string, opts coOpts) v1.Container {
	co := v1.Container{
		Name:      n,
		Image:     opts.image,
		Resources: v1.ResourceRequirements{},
	}

	if opts.rcpu != "" {
		co.Resources.Requests = makeRes(opts.rcpu, opts.rmem)
	}
	if opts.lcpu != "" {
		co.Resources.Limits = makeRes(opts.lcpu, opts.lmem)
	}
	if opts.lprob {
		co.LivenessProbe = &v1.Probe{}
	}
	if opts.rprob {
		co.ReadinessProbe = &v1.Probe{}
	}

	return co
}

func makeRes(c, m string) v1.ResourceList {
	return v1.ResourceList{
		v1.ResourceCPU:    *makeQty(c),
		v1.ResourceMemory: *makeQty(m),
	}
}

func makeQty(s string) *resource.Quantity {
	if s == "" {
		return nil
	}

	qty, _ := resource.ParseQuantity(s)
	return &qty
}
