package pkg

import "testing"

func TestParseBucket(t *testing.T) {
	var uu = []struct {
		bucketURI string
		bucket    string
		key       string
	}{
		{
			"s3://bucketName/",
			"bucketName",
			"",
		},
		{
			"bucket/with/subkey",
			"bucket",
			"with/subkey",
		},
		{
			"/bucket/with/leading/slashes/",
			"bucket",
			"with/leading/slashes",
		},
	}

	for _, v := range uu {
		u := v
		t.Run(u.bucketURI, func(t *testing.T) {
			b, k, err := parseBucket(u.bucketURI)
			if err != nil {
				t.Errorf("error got %v, want none", err)
			}
			if b != u.bucket {
				t.Errorf("bucket got %s, want %s", b, u.bucket)
			}
			if k != u.key {
				t.Errorf("key got %s, want %s", k, u.key)
			}
		})
	}
}

func TestParseBucketError(t *testing.T) {
	bucketURI := "s4://wrongbucket"
	_, _, err := parseBucket(bucketURI)
	if err != ErrUnknownS3BucketProtocol {
		t.Errorf("error expected %v, got none", ErrUnknownS3BucketProtocol)
	}
}
