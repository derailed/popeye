package cache

// AllKeys indicates all data keys are being used when referencing a cm or secret.
const AllKeys = "all"

// Blank represents an empty value.
var Blank = Empty{}

type (
	// ObjReferences tracks kubernetest object references.
	ObjReferences map[string]StringSet

	// Empty denotes an empty value.
	Empty struct{}

	// StringSet represents a set of strings.
	StringSet map[string]Empty
)

// Add a collection of elements to the set.
func (ss StringSet) Add(strs ...string) {
	for _, s := range strs {
		ss[s] = Blank
	}
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
