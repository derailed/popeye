// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package internal

import (
	"slices"

	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog/log"
)

type R string

var Glossary = make(Linters)

func init() {
	for _, r := range Rs {
		Glossary[r] = types.BlankGVR
	}
}

const (
	CM   R = "configmaps"
	CL   R = "cluster"
	EP   R = "endpoints"
	NS   R = "namespaces"
	NO   R = "nodes"
	PV   R = "persistentvolumes"
	PVC  R = "persistentvolumeclaims"
	PO   R = "pods"
	SEC  R = "secrets"
	SA   R = "serviceaccounts"
	SVC  R = "services"
	DP   R = "deployments"
	DS   R = "daemonsets"
	RS   R = "replicasets"
	STS  R = "statefulsets"
	CR   R = "clusterroles"
	CRB  R = "clusterrolebindings"
	RO   R = "roles"
	ROB  R = "rolebindings"
	ING  R = "ingresses"
	NP   R = "networkpolicies"
	PDB  R = "poddisruptionbudgets"
	HPA  R = "horizontalpodautoscalers"
	PMX  R = "podmetrics"
	NMX  R = "nodemetrics"
	CJOB R = "cronjobs"
	JOB  R = "jobs"
	GW   R = "gateways"
	GWC  R = "gatewayclasses"
	GWR  R = "httproutes"
)

var Rs = []R{
	CL, CM, EP, NS, NO, PV, PVC, PO, SEC, SA, SVC, DP, DS, RS, STS, CR,
	CRB, RO, ROB, ING, NP, PDB, HPA, PMX, NMX, CJOB, JOB, GW, GWC, GWR,
}

type Linters map[R]types.GVR

func (ll Linters) Dump() {
	log.Debug().Msg("\nLinters...")
	kk := make([]R, 0, len(ll))
	for k := range ll {
		kk = append(kk, k)
	}
	slices.Sort(kk)
	for _, k := range kk {
		log.Debug().Msgf("%-25s: %s", k, ll[k])
	}
}

func (ll Linters) Include(gvr types.GVR) (R, bool) {
	for k, v := range ll {
		g, r := v.G(), v.R()
		if g == gvr.G() && r == gvr.R() {
			return k, true
		}
	}

	return "", false
}
