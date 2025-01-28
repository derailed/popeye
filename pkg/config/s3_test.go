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
		kind   bucketKind
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
			kind:   s3Bucket,
			bucket: "bucketName",
		},

		"s3-toast": {
			uri: "s4://bucketName/",
			err: errors.New(`invalid S3 bucket URI: "s4://bucketName/"`),
		},

		"s3-with_full_key": {
			uri:    "s3://bucketName/fred/blee",
			bucket: "bucketName",
			kind:   s3Bucket,
			key:    "fred/blee",
		},

		"s3-with_key": {
			uri:    "bucket/with/subkey",
			bucket: "bucket",
			kind:   s3Bucket,
			key:    "with/subkey",
		},

		"s3-with_trailer": {
			uri:    "/bucket/with/leading/slashes/",
			bucket: "bucket",
			kind:   s3Bucket,
			key:    "with/leading/slashes",
		},

		"blee": {
			uri:    "my-bucket/popeye/my-cluster/2025/01/27/",
			bucket: "my-bucket",
			kind:   s3Bucket,
			key:    "popeye/my-cluster/2025/01/27",
		},

		"minio": {
			uri:    "minio://fred/blee/",
			bucket: "fred",
			kind:   minioBucket,
			key:    "blee",
		},

		"minio-with_key": {
			uri:    "minio://fred/blee/a/b.json",
			bucket: "fred",
			key:    "blee/a/b.json",
			kind:   minioBucket,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s3 := S3Info{Bucket: &u.uri}
			kind, b, k, err := s3.parse()

			assert.Equal(t, u.err, err)
			assert.Equal(t, u.kind, kind)
			assert.Equal(t, u.bucket, b)
			assert.Equal(t, u.key, k)
		})
	}
}
