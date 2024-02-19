// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package dag

import (
	"context"

	"github.com/Masterminds/semver"
)

// ListVersion return server api version.
func ListVersion(ctx context.Context) (*semver.Version, error) {
	f := mustExtractFactory(ctx)
	dial, err := f.Client().Dial()
	if err != nil {
		return nil, err
	}
	info, err := dial.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}
	rev, err := semver.NewVersion(info.Major + "." + info.Minor)
	if err != nil {
		return nil, err
	}

	return rev, nil
}
