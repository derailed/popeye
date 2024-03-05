// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package dag

import (
	"context"

	"github.com/blang/semver/v4"
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

	return ParseVersion(info)
}
