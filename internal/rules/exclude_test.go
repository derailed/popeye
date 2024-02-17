// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"testing"

	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func Test_excludesMatchGlobal(t *testing.T) {
	uu := map[string]struct {
		exc  Excludes
		spec Spec
		e    bool
	}{
		"empty": {},
		"empty-rule": {
			spec: Spec{
				GVR:  types.NewGVR("v1/pods"),
				FQN:  "ns1/p1",
				Code: 100,
			},
		},
		"happy-ns": {
			exc: Excludes{
				{
					FQNs: expressions{"rx:^ns1", "ns2"},
				},
			},
			spec: Spec{
				GVR:  types.NewGVR("v1/pods"),
				FQN:  "ns1/p1",
				Code: 100,
			},
			e: true,
		},
		"happy-labels": {
			exc: Excludes{
				{
					Labels: keyVals{
						"a": expressions{"b1", "b2"},
						"c": expressions{"d1", "d2"},
					},
				},
			},
			spec: Spec{
				GVR:    types.NewGVR("v1/pods"),
				FQN:    "ns3/p1",
				Labels: Labels{"a": "b1", "c": "d2"},
				Code:   100,
			},
			e: true,
		},
		"toast-labels": {
			exc: Excludes{
				{
					Labels: keyVals{
						"a": expressions{"b1", "b2"},
						"c": expressions{"d1", "d2"},
					},
				},
			},
			spec: Spec{
				GVR:    types.NewGVR("v1/pods"),
				FQN:    "ns3/p1",
				Labels: Labels{"a12": "b1", "c": "d12"},
				Code:   100,
			},
		},
		"happy-annotations": {
			exc: Excludes{
				{
					Annotations: keyVals{
						"a": expressions{"b1", "b2"},
						"c": expressions{"d1", "d2"},
					},
				},
			},
			spec: Spec{
				GVR:         types.NewGVR("v1/pods"),
				FQN:         "ns3/p1",
				Annotations: Labels{"a": "b1", "c": "d2"},
				Code:        100,
			},
			e: true,
		},
		"toast-annotations": {
			exc: Excludes{
				{
					Annotations: keyVals{
						"a": expressions{"b1", "b2"},
						"c": expressions{"d1", "d2"},
					},
				},
			},
			spec: Spec{
				GVR:         types.NewGVR("v1/pods"),
				FQN:         "ns3/p1",
				Annotations: Labels{"a": "b12", "c1": "d2"},
				Code:        100,
			},
		},
		"happy-co": {
			exc: Excludes{
				{
					Containers: expressions{"rx:^c"},
				},
			},
			spec: Spec{
				GVR:         types.NewGVR("v1/pods"),
				FQN:         "ns3/p1",
				Annotations: Labels{"a": "b1", "c": "d2"},
				Containers:  []string{"c1"},
				Code:        100,
			},
			e: true,
		},
		"toast-co": {
			exc: Excludes{
				{
					Containers: expressions{"rx:^c"},
				},
			},
			spec: Spec{
				GVR:         types.NewGVR("v1/pods"),
				FQN:         "ns3/p1",
				Annotations: Labels{"a": "b1", "c": "d2"},
				Containers:  []string{"fred"},
				Code:        100,
			},
		},
		"happy-code": {
			exc: Excludes{
				{
					Codes: expressions{"rx:^1"},
				},
			},
			spec: Spec{
				GVR:         types.NewGVR("v1/pods"),
				FQN:         "ns3/p1-code",
				Annotations: Labels{"a": "b1", "c": "d2"},
				Containers:  []string{"c1"},
				Code:        1666,
			},
			e: true,
		},
		"toast-code": {
			exc: Excludes{
				{
					Codes: expressions{"rx:^1"},
				},
			},
			spec: Spec{
				GVR:         types.NewGVR("v1/pods"),
				FQN:         "ns3/p1-code",
				Annotations: Labels{"a": "b1", "c": "d2"},
				Containers:  []string{"c1"},
				Code:        666,
			},
		},
		"toast": {
			exc: Excludes{
				{
					FQNs: expressions{"ns1", "ns2"},
				},
			},
			spec: Spec{
				GVR:  types.NewGVR("v1/pods"),
				FQN:  "ns3/p1",
				Code: 100,
			},
		},
		"toast-rx": {
			exc: Excludes{
				{
					FQNs: expressions{"ns1", "rx:.*ns2$"},
				},
			},
			spec: Spec{
				GVR:  types.NewGVR("v1/pods"),
				FQN:  "fred-ns2-blee/p1",
				Code: 100,
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			ok := u.exc.Match(u.spec, true)
			assert.Equal(t, u.e, ok)
		})
	}
}

func Test_excludesMatch(t *testing.T) {
	uu := map[string]struct {
		exc  Excludes
		spec Spec
		e    bool
	}{
		"empty": {},
		"empty-rule": {
			spec: Spec{
				GVR:  types.NewGVR("v1/pods"),
				FQN:  "ns1/p1",
				Code: 100,
			},
		},
		"happy": {
			exc: Excludes{
				{
					FQNs: expressions{"rx:^ns1", "ns2"},
				},
			},
			spec: Spec{
				GVR:  types.NewGVR("v1/pods"),
				FQN:  "ns1/p1",
				Code: 100,
			},
			e: true,
		},
		"happy-rx": {
			exc: Excludes{
				{
					FQNs: expressions{"ns1", "rx:.*ns2"},
				},
			},
			spec: Spec{
				GVR:  types.NewGVR("v1/pods"),
				FQN:  "fred-ns2/p1",
				Code: 100,
			},
			e: true,
		},
		"toast": {
			exc: Excludes{
				{
					FQNs: expressions{"ns1", "ns2"},
				},
			},
			spec: Spec{
				GVR:  types.NewGVR("v1/pods"),
				FQN:  "ns3/p1",
				Code: 100,
			},
		},
		"toast-rx": {
			exc: Excludes{
				{
					FQNs: expressions{"ns1", "rx:.*ns2$"},
				},
			},
			spec: Spec{
				GVR:  types.NewGVR("v1/pods"),
				FQN:  "fred-ns2-blee/p1",
				Code: 100,
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			ok := u.exc.Match(u.spec, false)
			assert.Equal(t, u.e, ok)
		})
	}
}

func Test_excludeMatch(t *testing.T) {
	uu := map[string]struct {
		exclude Exclude
		spec    Spec
		glob    bool
		e       bool
	}{
		"empty": {},
		"empty-rule": {
			spec: Spec{
				GVR:  types.NewGVR("v1/pods"),
				FQN:  "ns1/p1",
				Code: 100,
			},
		},
		"empty-spec": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
			},
		},

		"match-ns": {
			exclude: Exclude{
				FQNs: expressions{
					"rx:^ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
			},
			spec: Spec{
				FQN: "ns1/fred",
			},
			e: true,
		},
		"match-ns-rx": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
			},
			spec: Spec{
				FQN: "fred/blee",
			},
			e: true,
		},
		"match-ns-rx1": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
			},
			spec: Spec{
				FQN: "fred-blee/duh",
			},
			e: true,
		},
		"skip-ns-rx": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
			},
			spec: Spec{
				FQN: "zorg-bozo/duh",
			},
		},

		"match-labels-no-rule": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
			},
			spec: Spec{
				FQN:    "fred/duh",
				Labels: Labels{"a": "1", "b": "2"},
			},
			e: true,
		},
		"exact-match-labels": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"1"},
					"b": expressions{"2"},
				},
			},
			spec: Spec{
				FQN:    "fred/duh",
				Labels: Labels{"a": "1", "b": "2"},
			},
			e: true,
		},
		"set-match-labels": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"1", "2"},
					"b": expressions{"2", "3"},
				},
			},
			spec: Spec{
				FQN:    "fred/duh",
				Labels: Labels{"a": "1", "b": "2"},
			},
			e: true,
		},
		"match-labels-partial": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"1", "2"},
					"b": expressions{"2", "3"},
				},
			},
			spec: Spec{
				FQN:    "fred/duh",
				Labels: Labels{"a": "3", "b": "2"},
			},
			e: true,
		},
		"skip-labels-partial": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"1", "2"},
					"b": expressions{"2", "3"},
				},
			},
			spec: Spec{
				FQN:    "fred/duh",
				Labels: Labels{"a": "3", "b": "5"},
			},
		},
		"skip-labels-full": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"1", "2"},
					"b": expressions{"2", "3"},
				},
			},
			spec: Spec{
				FQN:    "fred/duh",
				Labels: Labels{"c": "3", "d": "5"},
			},
		},
		"match-labels-rx": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
			},
			spec: Spec{
				FQN:    "fred/duh",
				Labels: Labels{"a": "fred-duh", "b": "5"},
			},
			e: true,
		},
		"skip-labels-rx": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
			},
			spec: Spec{
				FQN:    "fred/duh",
				Labels: Labels{"a": "bozo-duh", "b": "5"},
			},
		},

		"skip-annot-rx": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Annotations: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
			},
			spec: Spec{
				FQN:         "fred/duh",
				Annotations: Labels{"a": "bozo-duh", "b": "5"},
			},
		},

		"match-container": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Annotations: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Containers: expressions{
					"c1",
					"c2",
					"rx:^fred",
				},
			},
			spec: Spec{
				FQN:         "fred/duh",
				Labels:      Labels{"a": "2", "b": "5"},
				Annotations: Labels{"a": "bozo-duh", "b": "3"},
				Containers:  []string{"c1"},
			},
			e: true,
		},
		"match-container-rx": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Annotations: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Containers: expressions{
					"c1",
					"c2",
					"rx:^fred",
				},
			},
			spec: Spec{
				FQN:         "fred/duh",
				Labels:      Labels{"a": "2", "b": "5"},
				Annotations: Labels{"a": "bozo-duh", "b": "3"},
				Containers:  []string{"fred-blee"},
			},
			e: true,
		},
		"skip-container-rx": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Annotations: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Containers: expressions{
					"c1",
					"c2",
					"rx:^fred",
				},
			},
			spec: Spec{
				FQN:         "fred/duh",
				Labels:      Labels{"a": "2", "b": "5"},
				Annotations: Labels{"a": "bozo-duh", "b": "3"},
				Containers:  []string{"blee-duh"},
			},
		},

		"match-all-codes": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Annotations: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Containers: expressions{
					"c1",
					"c2",
					"rx:^fred",
				},
			},
			spec: Spec{
				FQN:         "fred/duh",
				Labels:      Labels{"a": "2", "b": "5"},
				Annotations: Labels{"a": "bozo-duh", "b": "3"},
				Containers:  []string{"fred-blee"},
				Code:        100,
			},
			e: true,
		},
		"match-codes": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Annotations: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Containers: expressions{
					"c1",
					"c2",
					"rx:^fred",
				},
				Codes: expressions{
					"100",
					"200",
					"rx:^3",
				},
			},
			spec: Spec{
				FQN:         "fred/duh",
				Labels:      Labels{"a": "2", "b": "5"},
				Annotations: Labels{"a": "bozo-duh", "b": "3"},
				Containers:  []string{"fred-blee"},
				Code:        100,
			},
			e: true,
		},
		"match-code-rx": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Annotations: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Containers: expressions{
					"c1",
					"c2",
					"rx:^fred",
				},
				Codes: expressions{
					"100",
					"200",
					"rx:^3",
				},
			},
			spec: Spec{
				FQN:         "fred/duh",
				Labels:      Labels{"a": "2", "b": "5"},
				Annotations: Labels{"a": "bozo-duh", "b": "3"},
				Containers:  []string{"fred-blee"},
				Code:        333,
			},
			e: true,
		},
		"skip-code": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Annotations: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Containers: expressions{
					"c1",
					"c2",
					"rx:^fred",
				},
				Codes: expressions{
					"100",
					"102",
					"200",
				},
			},
			spec: Spec{
				FQN:         "fred/duh",
				Labels:      Labels{"a": "2", "b": "5"},
				Annotations: Labels{"a": "bozo-duh", "b": "3"},
				Containers:  []string{"fred-blee"},
				Code:        666,
			},
		},
		"skip-code-rx": {
			exclude: Exclude{
				FQNs: expressions{
					"ns1",
					"ns2",
					"rx:^fred",
					"rx:blee",
				},
				Labels: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Annotations: keyVals{
					"a": expressions{"rx:^fred", "2"},
					"b": expressions{"2", "3"},
				},
				Containers: expressions{
					"c1",
					"c2",
					"rx:^fred",
				},
				Codes: expressions{
					"100",
					"200",
					"rx:^3",
				},
			},
			spec: Spec{
				FQN:         "fred/duh",
				Labels:      Labels{"a": "2", "b": "5"},
				Annotations: Labels{"a": "bozo-duh", "b": "3"},
				Containers:  []string{"fred-blee"},
				Code:        633,
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			ok := u.exclude.Match(u.spec, u.glob)
			assert.Equal(t, u.e, ok)
		})
	}
}
