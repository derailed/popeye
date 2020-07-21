package pkg

import (
	"testing"
)

func TestParseBucket(t *testing.T) {
	var uu = []struct {
		bucketURI string
		bucket    string
		key       string
		endpoint  string
	}{
		{
			"minio://hostname:9000/bucketName/",
			"bucketName",
			"",
			"hostname",
		},
		{
			"minio://hostname:9000/bucketName/with/subkey/to/test",
			"bucketName",
			"with/subkey/to/test",
			"hostname",
		},
		{
			"s3://bucketName/",
			"bucketName",
			"",
			"",
		},
		{
			"bucket/with/subkey",
			"bucket",
			"with/subkey",
			"",
		},
		{
			"/bucket/with/leading/slashes/",
			"bucket",
			"with/leading/slashes",
			"",
		},
	}

	for _, v := range uu {
		u := v
		t.Run(u.bucketURI, func(t *testing.T) {
			h, b, k, err := parseBucket(u.bucketURI)
			if err != nil {
				t.Errorf("error got %v, want none", err)
			}
			if b != u.bucket {
				t.Errorf("bucket got %s, want %s", b, u.bucket)
			}
			if k != u.key {
				t.Errorf("key got %s, want %s", k, u.key)
			}
			if u.bucketURI == "minio://hostname" {
				if h != u.endpoint {
					t.Errorf("host is #{h}, want #{u.endpoint}")
				}
				if k != u.key {
					t.Errorf("key got %s, want %s", k, u.key)
				}
			}
		})
	}
}

func TestParseBucketError(t *testing.T) {
	bucketURI := "s4://wrongbucket"
	_, _, _, err := parseBucket(bucketURI)
	if err != ErrUnknownS3BucketProtocol {
		t.Errorf("error expected %v, got none", ErrUnknownS3BucketProtocol)
	}
}
