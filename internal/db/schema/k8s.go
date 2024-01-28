// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package schema

import (
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/types"
	"github.com/hashicorp/go-memdb"
)

// Init initializes db tables.
func Init() *memdb.DBSchema {
	var sc memdb.DBSchema
	sc.Tables = make(map[string]*memdb.TableSchema)
	for _, gvr := range internal.Glossary {
		if gvr == types.BlankGVR {
			continue
		}
		sc.Tables[gvr.String()] = indexFor(gvr.String())
	}

	return &sc
}
