// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import (
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cilium"
	iscrub "github.com/derailed/popeye/internal/scrub"
)

func Inject(ss map[internal.R]iscrub.ScrubFn) {
	ss[cilium.CEP] = NewCiliumEndpoint
	ss[cilium.CID] = NewCiliumIdentity
	ss[cilium.CNP] = NewCiliumNetworkPolicy
	ss[cilium.CCNP] = NewCiliumClusterwideNetworkPolicy
}
