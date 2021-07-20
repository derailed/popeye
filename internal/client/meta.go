package client

import (
	"github.com/derailed/popeye/types"
)

// Schema tracks resource schema.
type Schema struct {
	GVR       GVR
	Preferred bool
}

// Meta tracks a collection of resources.
type Meta map[string][]Schema

func newMeta() Meta {
	return make(map[string][]Schema)
}

// Resources tracks dictionary of resources.
var Resources = newMeta()

// Load loads resource meta from server.
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
