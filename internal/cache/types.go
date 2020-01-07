package cache

import (
	"github.com/derailed/popeye/internal"
)

// AllKeys indicates all data keys are being used when referencing a cm or secret.
const AllKeys = "all"

// ObjReferences tracks kubernetes object references.
type ObjReferences map[string]internal.StringSet
