package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBucket(t *testing.T) {
	var uu = map[string]struct {
		uri    string
		bucket string
		key    string
		err    error
	}{
		"s3_bucket": {
			uri:    "s3://bucketName/",
			bucket: "bucketName",
		},
		"toast": {
			uri: "s4://bucketName/",
			err: ErrUnknownS3BucketProtocol,
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
			b, k, err := parseBucket(u.uri)
			if err != u.err {
				t.Fatalf("error got %v, want none", err)
				return
			}
			assert.Equal(t, u.bucket, b)
			assert.Equal(t, u.key, k)
		})
	}
}
