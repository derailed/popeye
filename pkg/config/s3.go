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
	bucket, key, err := s.parse()
	if err != nil {
		return err
	}

	// Create a single AWS session (we can re use this if we're uploading many files)
	session, err := session.NewSession(&aws.Config{
		Logger:   &s3Logger{},
		LogLevel: aws.LogLevel(aws.LogDebugWithRequestErrors),
		Endpoint: s.Endpoint,
		Region:   s.Region,
	})
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

func (s *S3Info) parse() (string, string, error) {
	if !IsStrSet(s.Bucket) {
		return "", "", fmt.Errorf("invalid S3 bucket URI: %q", *s.Bucket)
	}
	u, err := url.Parse(*s.Bucket)
	if err != nil {
		return "", "", err
	}
	switch u.Scheme {
	// s3://bucket or s3://bucket/
	case "s3":
		var key string
		if u.Path != "" {
			key = strings.Trim(u.Path, "/")
		}
		return u.Host, key, nil
	// bucket/ or bucket/path/to/key
	case "":
		tokens := strings.SplitAfterN(strings.Trim(u.Path, "/"), "/", 2)
		if len(tokens) == 0 {
			return "", "", fmt.Errorf("invalid S3 bucket URI: %q", u.String())
		}
		key, bucket := "", strings.Trim(tokens[0], "/")
		if len(tokens) > 1 {
			key = tokens[1]
		}
		return bucket, key, nil
	default:
		return "", "", fmt.Errorf("invalid S3 bucket URI: %q", u.String())
	}
}

func newS3Info() *S3Info {
	return &S3Info{
		Bucket:   strPtr(""),
		Region:   strPtr(""),
		Endpoint: strPtr(""),
	}
}
