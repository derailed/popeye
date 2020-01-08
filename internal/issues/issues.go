package issues

import (
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
	var ii Issues
	for _, item := range o[section] {
		if item.Group != group {
			continue
		}
		ii = append(ii, item)
	}

	return ii
}
