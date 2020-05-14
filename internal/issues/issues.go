package issues

import (
	"sort"

	"github.com/derailed/popeye/pkg/config"
)

// Root denotes a root issue group.
const Root = "__root__"

type (
	// Issues represents a collection of issues.
	Issues []Issue

	// Outcome represents outcomes resulting from sanitization pass.
	Outcome map[string]Issues
)

// MaxSeverity gather the max severity in a collection of issues.
func (i Issues) MaxSeverity() config.Level {
	max := config.OkLevel
	for _, is := range i {
		if is.Level > max {
			max = is.Level
		}
	}

	return max
}

// Sort sorts issues.
func (i Issues) Sort(l config.Level) Issues {
	ii := make(Issues, 0, len(i))
	gg := i.Group()
	keys := make(sort.StringSlice, 0, len(gg))
	for k := range gg {
		keys = append(keys, k)
	}
	keys.Sort()
	for _, group := range keys {
		sev := gg[group].MaxSeverity()
		if sev < l {
			continue
		}
		for _, i := range gg[group] {
			if i.Level < l {
				continue
			}
			if i.Group == Root {
				ii = append(ii, i)
				continue
			}
			ii = append(ii, i)
		}
	}
	return ii
}

// Group collect issues as groups.
func (i Issues) Group() map[string]Issues {
	res := make(map[string]Issues)
	for _, item := range i {
		res[item.Group] = append(res[item.Group], item)
	}

	return res
}

// MaxSeverity scans the issues and reports the highest severity.
func (o Outcome) MaxSeverity(section string) config.Level {
	return o[section].MaxSeverity()
}

// MaxGroupSeverity scans the issues and reports the highest severity.
func (o Outcome) MaxGroupSeverity(section, group string) config.Level {
	return o.For(section, group).MaxSeverity()
}

// For returns issues for a given section/group.
func (o Outcome) For(section, group string) Issues {
	ii := make(Issues, 0, len(o[section]))
	for _, item := range o[section] {
		if item.Group != group {
			continue
		}
		ii = append(ii, item)
	}

	return ii
}

// Filter filters outcomes based on sanitizer level.
func (o Outcome) Filter(level config.Level) Outcome {
	for k, issues := range o {
		vv := make(Issues, 0, len(issues))
		for _, issue := range issues {
			if issue.Level >= level {
				vv = append(vv, issue)
			}
		}
		o[k] = vv
	}
	return o
}
