// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package dag

import (
	"context"
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
)

// ParseVersion renders cluster info version into semver rev.
func ParseVersion(info *version.Info) (*semver.Version, error) {
	if info == nil {
		return nil, fmt.Errorf("no cluster version available")
	}
	v := strings.TrimSuffix(info.Major+"."+info.Minor, "+")
	rev, err := semver.ParseTolerant(v)
	if err != nil {
		err = fmt.Errorf("semver parse failed for %q (%q|%q): %w", v, info.Major, info.Minor, err)
	}

	return &rev, err
}

func mustExtractFactory(ctx context.Context) types.Factory {
	f, ok := ctx.Value(internal.KeyFactory).(types.Factory)
	if !ok {
		panic("expecting factory in context")
	}
	return f
}

// MetaFQN returns a full qualified ns/name string.
func metaFQN(m metav1.ObjectMeta) string {
	if m.Namespace == "" {
		return m.Name
	}

	return m.Namespace + "/" + m.Name
}
