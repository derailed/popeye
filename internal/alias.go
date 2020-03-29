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

// ToResources converts aliases to resource names.
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
	a.aliases = map[string]Alias{
		"cluster":                 {ShortNames: StringSet{"cl": Blank}},
		"configmap":               {ShortNames: StringSet{"cm": Blank}},
		"clusterrole":             {ShortNames: StringSet{"cr": Blank}},
		"clusterrolebinding":      {ShortNames: StringSet{"crb": Blank}},
		"deployment":              {ShortNames: StringSet{"dp": Blank, "deploy": Blank}},
		"daemonset":               {ShortNames: StringSet{"ds": Blank}},
		"horizontalpodautoscaler": {ShortNames: StringSet{"hpa": Blank}},
		"ingress":                 {ShortNames: StringSet{"ing": Blank}, Plural: "ingresses"},
		"node":                    {ShortNames: StringSet{"no": Blank}},
		"networkpolicy":           {ShortNames: StringSet{"np": Blank}, Plural: "networkpolicies"},
		"namespace":               {ShortNames: StringSet{"ns": Blank}},
		"poddisruptionbudget":     {ShortNames: StringSet{"pdb": Blank}},
		"pod":                     {ShortNames: StringSet{"po": Blank}},
		"podsecuritypolicy":       {ShortNames: StringSet{"psp": Blank}, Plural: "podsecuritypolicies"},
		"persistentvolume":        {ShortNames: StringSet{"pv": Blank}},
		"persistentvolumeclaim":   {ShortNames: StringSet{"pvc": Blank}},
		"rolebinding":             {ShortNames: StringSet{"rb": Blank}},
		"role":                    {ShortNames: StringSet{"ro": Blank}},
		"replicaset":              {ShortNames: StringSet{"rs": Blank}},
		"serviceaccount":          {ShortNames: StringSet{"sa": Blank}},
		"secret":                  {ShortNames: StringSet{"sec": Blank}},
		"statefulset":             {ShortNames: StringSet{"sts": Blank}},
		"service":                 {ShortNames: StringSet{"svc": Blank}},
	}
}
