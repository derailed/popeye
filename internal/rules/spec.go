// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"fmt"
	"strings"

	"github.com/derailed/popeye/types"
)

// Spec tracks an issue spec
type Spec struct {
	GVR         types.GVR
	FQN         string
	Labels      Labels
	Annotations Labels
	Containers  []string
	Code        ID
}

func (s Spec) isEmpty() bool {
	return s.GVR == types.BlankGVR &&
		s.FQN == "" &&
		len(s.Labels) == 0 &&
		len(s.Annotations) == 0 &&
		len(s.Containers) == 0 &&
		s.Code == ZeroCode
}

func (s Spec) String() string {
	ss := fmt.Sprintf("[%s] %s", s.GVR, s.FQN)
	if len(s.Containers) != 0 {
		ss += fmt.Sprintf("::%s", strings.Join(s.Containers, ","))
	}
	if s.Code != ZeroCode {
		ss += fmt.Sprintf("(%q)", s.Code)
	}
	ss += fmt.Sprintf("-- %s::%s", s.Labels, s.Annotations)

	return ss
}
