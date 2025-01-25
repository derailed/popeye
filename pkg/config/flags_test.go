// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPersistent(t *testing.T) {
	uu := map[string]struct {
		f Flags
		e bool
	}{
		"empty": {},
		"blank": {
			f: Flags{},
		},
		"save": {
			f: Flags{
				Save: boolPtr(true),
			},
			e: true,
		},
		"s3": {
			f: Flags{
				S3: &S3Info{
					Bucket: strPtr("blah"),
				},
			},
			e: true,
		},
		"output-file": {
			f: Flags{
				OutputFile: strPtr("blah"),
			},
			e: true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.f.IsPersistent())
		})
	}
}

func TestExhaust(t *testing.T) {
	uu := map[string]struct {
		f Flags
		e string
	}{
		"empty": {},
		"blank": {
			f: Flags{},
		},
		"save": {
			f: Flags{
				Save: boolPtr(true),
			},
		},
		"s3": {
			f: Flags{
				S3: &S3Info{
					Bucket: strPtr("blah"),
				},
			},
			e: "blah",
		},
		"output-file": {
			f: Flags{
				OutputFile: strPtr("blah"),
			},
			e: "blah",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.f.Exhaust())
		})
	}
}

func TestOutputFormat(t *testing.T) {
	uu := map[string]struct {
		f Flags
		e string
	}{
		"standard": {Flags{Output: strPtr("standard")}, "standard"},
		"blank":    {Flags{Output: strPtr("")}, "cool"},
		"nil":      {Flags{}, "cool"},
		"blee":     {Flags{Output: strPtr("blee")}, "blee"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.f.OutputFormat())
		})
	}
}
