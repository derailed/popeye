// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBucket(t *testing.T) {
	var uu = map[string]struct {
		uri    string
		host   string
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
		"s3-toast": {
			uri: "s4://bucketName/",
			err: errors.New(`invalid S3 bucket URI: "s4://bucketName/"`),
		},
		"s3-with_full_key": {
			uri:    "s3://bucketName/fred/blee",
			bucket: "bucketName",
			key:    "fred/blee",
		},
		"s3-with_key": {
			uri:    "bucket/with/subkey",
			bucket: "bucket",
			key:    "with/subkey",
		},
		"s3-with_trailer": {
			uri:    "/bucket/with/leading/slashes/",
			bucket: "bucket",
			key:    "with/leading/slashes",
		},
		"minio": {
			uri:    "minio://hostname:9000/bucketName/",
			bucket: "bucketName",
			host:   "hostname:9000",
		},
		"minio-with_key": {
			uri:    "minio://hostname:9000/bucketName/with/subkey/to/test",
			bucket: "bucketName",
			key:    "with/subkey/to/test",
			host:   "hostname:9000",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s3 := S3Info{Bucket: &u.uri}
			h, b, k, err := s3.parse()

			assert.Equal(t, u.err, err)
			assert.Equal(t, u.host, h)
			assert.Equal(t, u.bucket, b)
			assert.Equal(t, u.key, k)
		})
	}
}
