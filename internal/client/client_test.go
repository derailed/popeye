// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSet(t *testing.T) {
	s1, s2 := "fred", ""

	uu := []struct {
		s *string
		e bool
	}{
		{&s1, true},
		{&s2, false},
		{nil, false},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, isSet(u.s))
	}
}
