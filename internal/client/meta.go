package client

import (
	"github.com/derailed/popeye/types"
)

type Schema struct {
	GVR       GVR
	Preferred bool
}

type Meta map[string][]Schema

func newMeta() Meta {
	return make(map[string][]Schema)
}

var Resources = newMeta()

func Load(f types.Factory) error {
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
			gvr := FromGVAndR(r.GroupVersion, res.Name)
			res.Group, res.Version = gvr.G(), gvr.V()
			Resources[gvr.R()] = []Schema{{GVR: gvr, Preferred: true}}
		}
	}

	return nil
}
