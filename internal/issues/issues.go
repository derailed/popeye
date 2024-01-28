// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package issues

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues/tally"
	"github.com/derailed/popeye/internal/rules"
)

// Root denotes a root issue group.
const Root = "__root__"

type (
	// Issues represents a collection of issues.
	Issues []Issue

	// Outcome represents outcomes resulting from sanitization pass.
	Outcome map[string]Issues
)

func (i Issues) CodeTally() tally.Code {
	ss := make(tally.Code)
	for _, issue := range i {
		if c, ok := issue.Code(); ok {
			if v, ok := ss[c]; ok {
				ss[c] = v + 1
			} else {
				ss[c] = 1
			}
		}
	}

	return ss
}

// MaxSeverity gather the max severity in a collection of issues.
func (i Issues) MaxSeverity() rules.Level {
	max := rules.OkLevel
	for _, is := range i {
		if is.Level > max {
			max = is.Level
		}
	}

	return max
}

func (i Issues) HasIssues() bool {
	return len(i) > 0
}

func SortKeys(k1, k2 string) int {
	v1, err := strconv.Atoi(k1)
	if err == nil {
		v2, _ := strconv.Atoi(k2)
		switch {
		case v1 == v2:
			return 0
		case v1 < v2:
			return -1
		default:
			return 1
		}
	}

	return strings.Compare(k1, k2)
}

// Sort sorts issues.
func (i Issues) Sort(l rules.Level) Issues {
	ii := make(Issues, 0, len(i))
	gg := i.Group()
	kk := make([]string, 0, len(gg))
	for k := range gg {
		kk = append(kk, k)
	}
	slices.SortFunc(kk, SortKeys)

	for _, k := range kk {
		ii = append(ii, gg[k]...)
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

// NSTally collects Namespace code tally for a given linter.
func (o Outcome) NSTally() tally.Namespace {
	nn := make(tally.Namespace, len(o))
	for fqn, v := range o {
		ns, _ := client.Namespaced(fqn)
		if ns == "" {
			ns = "-"
		}
		tt := v.CodeTally()
		if v1, ok := nn[ns]; ok {
			v1.Merge(tt)
		} else {
			nn[ns] = tt
		}
	}

	return nn
}

// MaxSeverity scans the issues and reports the highest severity.
func (o Outcome) MaxSeverity(section string) rules.Level {
	return o[section].MaxSeverity()
}

func (o Outcome) MarshalJSON() ([]byte, error) {
	out := make([]string, 0, len(o))
	for k, v := range o {
		if len(v) == 0 {
			continue
		}
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		out = append(out, fmt.Sprintf("%q: %s", k, raw))
	}

	return []byte("{" + strings.Join(out, ",") + "}"), nil
}

func (o Outcome) MarshalYAML() (interface{}, error) {
	out := make(Outcome, len(o))
	for k, v := range o {
		if len(v) == 0 {
			continue
		}
		out[k] = v
	}

	return out, nil
}

func (o Outcome) HasIssues() bool {
	var count int
	for _, ii := range o {
		count += len(ii)
	}

	return count > 0
}

// MaxGroupSeverity scans the issues and reports the highest severity.
func (o Outcome) MaxGroupSeverity(section, group string) rules.Level {
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

// Filter filters outcomes based on lint level.
func (o Outcome) Filter(level rules.Level) Outcome {
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

func (o Outcome) Dump() {
	if len(o) == 0 {
		fmt.Println("No ISSUES!")
	}
	for k, ii := range o {
		fmt.Println(k)
		for _, i := range ii {
			i.Dump()
		}
	}
}
