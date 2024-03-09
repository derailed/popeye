// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cilium

import (
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/types"
)

func init() {
	for _, r := range CiliumRS {
		internal.Glossary[r] = types.BlankGVR
	}
}

const (
	CEP  internal.R = "ciliumendpoints"
	CID  internal.R = "ciliumidentities"
	CNP  internal.R = "ciliumnetworkpolicies"
	CCNP internal.R = "ciliumclusterwidenetworkpolicies"
)

var CiliumRS = []internal.R{CEP, CID, CNP, CCNP}

var Aliases = internal.ShortNames{
	CEP: {"cep"},
	CID: {"cid"},
}
