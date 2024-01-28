// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBoolSet(t *testing.T) {
	true, false := true, false
	uu := map[string]struct {
		b *bool
		e bool
	}{
		"empty": {},
		"happy": {
			b: &true,
			e: true,
		},
		"false": {
			b: &false,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, IsBoolSet(u.b))
		})
	}
}

func TestIsStrSet(t *testing.T) {
	uu := map[string]struct {
		s *string
		e bool
	}{
		"empty": {},
		"happy": {
			s: strPtr("fred"),
			e: true,
		},
		"blank": {
			s: strPtr(""),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, IsStrSet(u.s))
		})
	}
}

func TestSanitizeFile(t *testing.T) {
	uu := map[string]struct {
		f, e string
	}{
		"empty": {},
		"plain": {
			f: "fred-bozo",
			e: "fred-bozo",
		},
		"full": {
			f: "fred/blee///duh::bozo",
			e: "fred-blee-duh-bozo",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, SanitizeFileName(u.f))
		})
	}
}
