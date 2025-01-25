// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rs/zerolog/log"
)

type s3Logger struct{}

func (s *s3Logger) Log(mm ...any) {
	for _, m := range mm {
		log.Debug().Msgf("S3 %s", m)
	}
}

type S3Info struct {
	Bucket   *string
	Region   *string
	Endpoint *string
}

func (s *S3Info) Upload(asset string, contentType string, rwc io.ReadWriteCloser) error {
	defer rwc.Close()

	log.Debug().Msgf("S3 bucket path: %q", asset)
	host, bucket, key, err := s.parse()
	if err != nil {
		return err
	}

	cfg := aws.Config{
		Logger:   &s3Logger{},
		LogLevel: aws.LogLevel(aws.LogDebugWithRequestErrors),
		Region:   s.Region,
		Endpoint: s.Endpoint,
	}

	if host != "" {
		cfg.Endpoint = aws.String(host)
		cfg.S3ForcePathStyle = aws.Bool(true)
		cfg.DisableSSL = aws.Bool(true)
	}

	// Create a single AWS session (we can re use this if we're uploading many files)
	session, err := session.NewSession(&cfg)
	if err != nil {
		return err
	}

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(session)
	// Upload input parameters
	upParams := s3manager.UploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(filepath.Join(key, asset)),
		Body:     rwc,
		Metadata: aws.StringMap(map[string]string{"Content-Type": contentType}),
	}
	// Perform an upload.
	if _, err = uploader.Upload(&upParams); err != nil {
		return err
	}

	return nil
}

func (s *S3Info) parse() (string, string, string, error) {
	if !IsStrSet(s.Bucket) {
		return "", "", "", fmt.Errorf("invalid S3 bucket URI: %q", *s.Bucket)
	}
	u, err := url.Parse(*s.Bucket)
	if err != nil {
		return "", "", "", err
	}
	switch u.Scheme {
	case "s3":
		// s3://bucket or s3://bucket/
		var key string
		if u.Path != "" {
			key = strings.Trim(u.Path, "/")
		}
		return "", u.Host, key, nil
	case "minio":
		var key, bucket string
		if u.Path != "" {
			bucketpath := strings.SplitN(u.Path[1:], "/", 2)
			bucket = bucketpath[0]
			key = bucketpath[1]
		}
		return u.Host, bucket, key, nil
	case "":
		// bucket/ or bucket/path/to/key
		tokens := strings.SplitAfterN(strings.Trim(u.Path, "/"), "/", 2)
		if len(tokens) == 0 {
			return "", "", "", fmt.Errorf("invalid S3 bucket URI: %q", u.String())
		}
		key, bucket := "", strings.Trim(tokens[0], "/")
		if len(tokens) > 1 {
			key = tokens[1]
		}
		return u.Host, bucket, key, nil
	default:
		return "", "", "", fmt.Errorf("invalid S3 bucket URI: %q", u.String())
	}
}

func newS3Info() *S3Info {
	return &S3Info{
		Bucket:   strPtr(""),
		Region:   strPtr(""),
		Endpoint: strPtr(""),
	}
}
