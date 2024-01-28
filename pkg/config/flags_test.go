// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

import (
	"errors"
	"net/url"
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

func TestParseBucket(t *testing.T) {
	var uu = map[string]struct {
		uri    string
		bucket string
		key    string
		err    error
	}{
		"empty": {
			err: errors.New(`invalid S3 bucket URI: ""`),
		},
		"no-scheme": {
			uri: ":bozo",
			err: &url.Error{Op: "parse", URL: ":bozo", Err: errors.New("missing protocol scheme")},
		},
		"s3_bucket": {
			uri:    "s3://bucketName/",
			bucket: "bucketName",
		},
		"toast": {
			uri: "s4://bucketName/",
			err: errors.New(`invalid S3 bucket URI: "s4://bucketName/"`),
		},
		"with_full_key": {
			uri:    "s3://bucketName/fred/blee",
			bucket: "bucketName",
			key:    "fred/blee",
		},
		"with_key": {
			uri:    "bucket/with/subkey",
			bucket: "bucket",
			key:    "with/subkey",
		},
		"with_trailer": {
			uri:    "/bucket/with/leading/slashes/",
			bucket: "bucket",
			key:    "with/leading/slashes",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s3 := S3Info{Bucket: &u.uri}
			b, k, err := s3.parse()

			assert.Equal(t, u.err, err)
			assert.Equal(t, u.bucket, b)
			assert.Equal(t, u.key, k)
		})
	}
}
