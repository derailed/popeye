package internal

import (
	"fmt"
	"strings"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceMetas represents a collection of resource metadata.
type ResourceMetas map[client.GVR]metav1.APIResource

// Aliases represents a collection of resource aliases.
type Aliases struct {
	aliases map[string]client.GVR
	metas   ResourceMetas
}

// NewAliases returns a new instance.
func NewAliases() *Aliases {
	a := Aliases{
		aliases: make(map[string]client.GVR),
		metas:   make(ResourceMetas),
	}

	return &a
}

// Init loads the aliases glossary.
func (a *Aliases) Init(f types.Factory, gvrs []string) error {
	if err := a.loadPreferred(f); err != nil {
		return err
	}
	for _, k := range gvrs {
		gvr := client.NewGVR(k)
		res, ok := a.metas[gvr]
		if !ok {
			panic(fmt.Sprintf("No resource meta found for %s", gvr))
		}
		a.aliases[res.Name] = gvr
		a.aliases[res.SingularName] = gvr
		for _, n := range res.ShortNames {
			a.aliases[n] = gvr
		}
	}
	a.aliases["cl"] = client.NewGVR("cluster")
	a.aliases["sec"] = client.NewGVR("v1/secrets")
	a.aliases["dp"] = client.NewGVR("apps/v1/deployments")
	a.aliases["cr"] = client.NewGVR("rbac.authorization.k8s.io/v1/clusterroles")
	a.aliases["crb"] = client.NewGVR("rbac.authorization.k8s.io/v1/clusterrolebindings")
	a.aliases["ro"] = client.NewGVR("rbac.authorization.k8s.io/v1/roles")
	a.aliases["rb"] = client.NewGVR("rbac.authorization.k8s.io/v1/rolebindings")
	a.aliases["np"] = client.NewGVR("networking.k8s.io/v1/networkpolicies")

	return nil
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

func (a *Aliases) loadPreferred(f types.Factory) error {
	dial, err := f.Client().CachedDiscovery()
	if err != nil {
		return err
	}
	rr, err := dial.ServerPreferredResources()
	if err != nil {
		return err
	}
	for _, r := range rr {
		for _, res := range r.APIResources {
			gvr := client.FromGVAndR(r.GroupVersion, res.Name)
			res.Group, res.Version = gvr.G(), gvr.V()
			if res.SingularName == "" {
				res.SingularName = strings.ToLower(res.Kind)
			}
			a.metas[gvr] = res
		}
	}

	a.metas[client.NewGVR("cluster")] = metav1.APIResource{
		Name: "cluster",
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
func (a *Aliases) Singular(gvr client.GVR) string {
	m, ok := a.metas[gvr]
	if !ok {
		log.Error().Msgf("Missing meta for gvr %q", gvr)
		return gvr.R()
	}
	return m.SingularName
}

// Exclude checks if section should be excluded from the report.
func (a *Aliases) Exclude(gvr client.GVR, sections []string) bool {
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
