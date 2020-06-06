package pkg

import "testing"

func TestParseBucket(t *testing.T) {
	var tests = []struct {
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

	for _, tt := range tests {
		t.Run(tt.bucketURI, func(t *testing.T) {
			b, k, err := parseBucket(tt.bucketURI)
			if err != nil {
				t.Errorf("error got %v, want none", err)
			}
			if b != tt.bucket {
				t.Errorf("bucket got %s, want %s", b, tt.bucket)
			}
			if k != tt.key {
				t.Errorf("key got %s, want %s", k, tt.key)
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
