// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package internal

// All indicates all data keys are being used when referencing a cm or secret.
const All = "all"

// Empty denotes an empty value.
type Empty struct{}

// Blank represents an empty value.
var Blank = Empty{}

// StringSet represents a set of strings.
type StringSet map[string]Empty

// AllKeys indicates all keys are present.
var AllKeys = StringSet{All: Blank}

// AddAll merges two sets.
func (ss StringSet) AddAll(s StringSet) {
	for k := range s {
		ss[k] = Blank
	}
}

// Add a collection of elements to the set.
func (ss StringSet) Add(strs ...string) {
	for _, s := range strs {
		ss[s] = Blank
	}
}

// Clone returns a new copy.
func (ss StringSet) Clone() StringSet {
	clone := make(StringSet, len(ss))
	for k, v := range ss {
		clone[k] = v
	}
	return clone
}

// Has checks if an item is in the set.
func (ss StringSet) Has(s string) bool {
	_, ok := ss[s]

	return ok
}

// Diff computes B-A.
func (ss StringSet) Diff(set StringSet) StringSet {
	delta := make(StringSet)
	for k := range set {
		if _, ok := ss[k]; !ok {
			delta.Add(k)
		}
	}

	return delta
}
