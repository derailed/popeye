package internal

// BOZO!! Canned for now - make k8s call for these and refine.

// Alias represents a resource alias.
type Alias struct {
	ShortNames StringSet
	Plural     string
}

// Aliases represents a collection of aliases.
type Aliases struct {
	aliases map[string]Alias
}

// NewAliases returns a new alias glossary.
func NewAliases() *Aliases {
	a := Aliases{}
	a.init()

	return &a
}

func (a *Aliases) ToResources(ss []string) []string {
	aa := make([]string, len(ss))
	for i := 0; i < len(ss); i++ {
		aa[i] = a.FromAlias(ss[i])
	}
	return aa
}

// Pluralize returns a plural form.
func (a Aliases) Pluralize(res string) string {
	if v, ok := a.aliases[res]; ok {
		if v.Plural != "" {
			return v.Plural
		}
	}
	return res + "s"
}

// FromAlias returns the resource name from an alias.
func (a Aliases) FromAlias(res string) string {
	if _, ok := a.aliases[res]; ok {
		return res
	}

	for k, v := range a.aliases {
		if _, ok := v.ShortNames[res]; ok {
			return k
		}
	}

	return res
}

func (a *Aliases) init() {
	// Glossary stores a collection of resource aliases.
	a.aliases = map[string]Alias{
		"cluster":                 Alias{ShortNames: StringSet{"cl": Blank}},
		"configmap":               Alias{ShortNames: StringSet{"cm": Blank}},
		"clusterrole":             Alias{ShortNames: StringSet{"cr": Blank}},
		"clusterrolebinding":      Alias{ShortNames: StringSet{"crb": Blank}},
		"deployment":              Alias{ShortNames: StringSet{"dp": Blank, "deploy": Blank}, Plural: "deployments"},
		"daemonset":               Alias{ShortNames: StringSet{"ds": Blank}},
		"horizontalpodautoscaler": Alias{ShortNames: StringSet{"hpa": Blank}},
		"ingress":                 Alias{ShortNames: StringSet{"ing": Blank}, Plural: "ingresses"},
		"node":                    Alias{ShortNames: StringSet{"no": Blank}},
		"networkpolicy":           Alias{ShortNames: StringSet{"np": Blank}, Plural: "networkpolicies"},
		"namespace":               Alias{ShortNames: StringSet{"ns": Blank}},
		"poddisruptionbudget":     Alias{ShortNames: StringSet{"pdb": Blank}},
		"pod":                     Alias{ShortNames: StringSet{"po": Blank}},
		"podsecuritypolicy":       Alias{ShortNames: StringSet{"psp": Blank}, Plural: "podsecuritypolicies"},
		"persistentvolume":        Alias{ShortNames: StringSet{"pv": Blank}},
		"persistentvolumeclaim":   Alias{ShortNames: StringSet{"pvc": Blank}},
		"rolebinding":             Alias{ShortNames: StringSet{"rb": Blank}},
		"role":                    Alias{ShortNames: StringSet{"ro": Blank}},
		"replicaset":              Alias{ShortNames: StringSet{"rs": Blank}},
		"serviceaccount":          Alias{ShortNames: StringSet{"sa": Blank}},
		"secret":                  Alias{ShortNames: StringSet{"sec": Blank}},
		"statefulset":             Alias{ShortNames: StringSet{"sts": Blank}},
		"service":                 Alias{ShortNames: StringSet{"svc": Blank}},
	}
}
