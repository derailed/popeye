// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package internal

import (
	"fmt"
	"slices"
	"strings"

	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	ClusterGVR = types.NewGVR("cluster")
	groups     = map[R]string{
		NP: "networking.k8s.io",
	}
)

// ResourceMetas represents a collection of resource metadata.
type ResourceMetas map[types.GVR]metav1.APIResource

// Aliases represents a collection of resource aliases.
type Aliases struct {
	aliases map[string]types.GVR
	metas   ResourceMetas
	cilium  bool
}

// NewAliases returns a new instance.
func NewAliases() *Aliases {
	a := Aliases{
		aliases: make(map[string]types.GVR),
		metas:   make(ResourceMetas),
	}

	return &a
}

func (a *Aliases) Dump() {
	log.Debug().Msgf("\nAliases...")
	kk := make([]string, 0, len(a.aliases))
	for k := range a.aliases {
		kk = append(kk, k)
	}
	slices.Sort(kk)
	for _, k := range kk {
		log.Debug().Msgf("%-25s: %s", k, a.aliases[k])
	}
}

type ShortNames map[R][]string

var customShortNames = ShortNames{
	CL:  {"cl"},
	SEC: {"sec"},
	DP:  {"dp"},
	CR:  {"cr"},
	CRB: {"crb"},
	RO:  {"ro"},
	ROB: {"rb"},
	NP:  {"np"},
	GWR: {"gwr"},
	GWC: {"gwc"},
	GW:  {"gw"},
}

func (a *Aliases) Inject(ss ShortNames) {
	for gvr, res := range a.metas {
		if kk, ok := ss[R(res.Name)]; ok {
			for _, k := range kk {
				a.aliases[k] = gvr
			}
		}
	}
}

func (a *Aliases) IsNamespaced(gvr types.GVR) bool {
	if r, ok := a.metas[gvr]; ok {
		return r.Namespaced
	}

	return true
}

// Init loads the aliases glossary.
func (a *Aliases) Init(c types.Connection) error {
	return a.loadPreferred(c)
}

func (a *Aliases) Realize() {
	for gvr, res := range a.metas {
		a.aliases[res.Name] = gvr
		if res.SingularName != "" {
			a.aliases[res.SingularName] = gvr
		}
		for _, n := range res.ShortNames {
			a.aliases[n] = gvr
		}
		if kk, ok := customShortNames[R(res.Name)]; ok {
			for _, k := range kk {
				a.aliases[k] = gvr
			}
		}
		if lgvr, ok := Glossary[R(res.SingularName)]; ok {
			if greaterV(gvr.V(), lgvr.V()) {
				Glossary[R(res.SingularName)] = gvr
			}
		} else if lgvr, ok := Glossary[R(res.Name)]; ok {
			if greaterV(gvr.V(), lgvr.V()) {
				Glossary[R(res.Name)] = gvr
			}
		}
	}
}

func greaterV(v1, v2 string) bool {
	if v1 == "" && v2 == "" {
		return true
	}
	if v2 == "" {
		return true
	}
	if v1 == "v1" || v1 == "v2" {
		return true
	}

	return false
}

// TitleFor produces a section title from an alias.
func (a *Aliases) TitleFor(s string, plural bool) string {
	gvr, ok := a.aliases[s]
	if !ok {
		panic(fmt.Sprintf("No alias for %q", s))
	}
	m, ok := a.metas[gvr]
	if !ok {
		panic(fmt.Sprintf("No meta for %q", gvr))
	}
	if plural {
		return m.Name
	}

	return m.SingularName
}

func (a *Aliases) IsCiliumCluster() bool {
	return a.cilium
}

func (a *Aliases) loadPreferred(c types.Connection) error {
	dial, err := c.CachedDiscovery()
	if err != nil {
		return err
	}
	ll, err := dial.ServerPreferredResources()
	if err != nil {
		return err
	}
	for _, l := range ll {
		gv, err := schema.ParseGroupVersion(l.GroupVersion)
		if err != nil {
			continue
		}
		for _, r := range l.APIResources {
			if g, ok := groups[R(r.Name)]; ok && g != gv.Group {
				continue
			}

			gvr := types.NewGVRFromAPIRes(gv, r)
			if !a.cilium && strings.Contains(gvr.G(), "cilium.io") {
				a.cilium = true
			}
			r.Group, r.Version = gvr.G(), gvr.V()
			if r.SingularName == "" {
				r.SingularName = strings.ToLower(r.Kind)
			}
			a.metas[gvr] = r
		}
	}
	a.metas[ClusterGVR] = metav1.APIResource{
		Name: ClusterGVR.String(),
	}

	return nil
}

// ToResources converts aliases to resource names.
func (a *Aliases) ToResources(nn []string) []string {
	rr := make([]string, 0, len(nn))
	for _, n := range nn {
		if gvr, ok := a.aliases[n]; ok {
			rr = append(rr, gvr.R())
		} else {
			panic(fmt.Sprintf("no aliases for %q", n))
		}
	}
	return rr
}

// Singular returns a singular resource name.
func (a *Aliases) Singular(gvr types.GVR) string {
	m, ok := a.metas[gvr]
	if !ok {
		log.Error().Msgf("Missing meta for gvr %q", gvr)
		return gvr.R()
	}

	return m.SingularName
}

// Exclude checks if section should be excluded from the report.
func (a *Aliases) Exclude(gvr types.GVR, sections []string) bool {
	if len(sections) == 0 {
		return false
	}

	var matches int
	for _, s := range sections {
		agvr, ok := a.aliases[s]
		if !ok {
			continue
		}
		if agvr.String() == gvr.String() {
			matches++
		}
	}

	return matches == 0
}
