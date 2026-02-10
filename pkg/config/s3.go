// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/logging"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

type bucketKind string

const (
	s3Bucket    bucketKind = "s3"
	minioBucket bucketKind = "minio"
)

type s3Logger struct{}

func (s *s3Logger) Log(mm ...any) {
	for _, m := range mm {
		log.Debug().Msgf("S3 %s", m)
	}
}

func (s *s3Logger) Logf(classification logging.Classification, format string, v ...interface{}) {
	log.Debug().Msgf("[AWS] %s", v)
}

type S3Info struct {
	Bucket   *string
	Region   *string
	Endpoint *string
}

func (s *S3Info) Upload(ctx context.Context, asset string, contentType string, rwc io.ReadWriteCloser) error {
	defer rwc.Close()

	log.Debug().Msgf("S3 bucket path: %q", asset)
	kind, bucket, key, err := s.parse()
	if err != nil {
		return err
	}

	switch kind {
	case s3Bucket:
		return s.awsUpload(ctx, bucket, key, asset, rwc)
	case minioBucket:
		return s.minioUpload(ctx, bucket, key, asset, rwc)
	default:
		return fmt.Errorf("unsupported S3 storage: %s", kind)
	}
}

func (s *S3Info) minioUpload(ctx context.Context, bucket, key, asset string, rwc io.ReadWriteCloser) error {
	minioClient, err := minio.New(*s.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			""),
		Secure: false,
	})
	if err != nil {
		return err
	}

	err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: *s.Region})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucket)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucket)
		} else {
			return err
		}
	}

	contentType := "application/octet-stream"
	info, err := minioClient.PutObject(ctx,
		bucket,
		filepath.Join(key, asset),
		rwc,
		-1,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return err
	}

	log.Info().Msgf("successfully uploaded %s of size %d\n", key, info.Size)
	return nil
}

func (s *S3Info) awsUpload(ctx context.Context, bucket, key, asset string, rwc io.ReadWriteCloser) error {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(*s.Region),
		config.WithLogConfigurationWarnings(true),
		config.WithLogger(&s3Logger{}),
	)
	if err != nil {
		return err
	}

	clt := s3.NewFromConfig(cfg)
	_, err = clt.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &bucket})
	if err != nil {
		log.Debug().Msgf("bucket %s not found or not accessible, attempting to create", bucket)
		opts := s3.CreateBucketInput{
			Bucket: &bucket,
			CreateBucketConfiguration: &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraint(*s.Region),
			},
		}
		if _, err = clt.CreateBucket(ctx, &opts); err != nil {
			var (
				exists *types.BucketAlreadyExists
				owned  *types.BucketAlreadyOwnedByYou
			)
			switch {
			case errors.As(err, &exists):
				log.Info().Msgf("bucket %s already exists", bucket)
			case errors.As(err, &owned):
				log.Info().Msgf("bucket %s already owned by you", bucket)
			default:
				log.Err(err).Msgf("failed to create bucket %s", bucket)
				return err
			}
		} else {
			log.Info().Msgf("successfully created bucket %s", bucket)
 		}
	} else {
		log.Info().Msgf("bucket %s already exists", bucket)
	}

	uploader := manager.NewUploader(clt)
	path := filepath.Join(key, asset)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &path,
		Body:   rwc,
	})
	if err != nil {
		log.Err(err).Msgf("failed to upload to bucket: %s//%s", bucket, path)
	} else {
		log.Info().Msgf("Success: uploaded to bucket: %s//%s", bucket, path)
	}

	return err
}

func (s *S3Info) parse() (bucketKind, string, string, error) {
	if !IsStrSet(s.Bucket) {
		return "", "", "", fmt.Errorf("invalid S3 bucket URI: %q", *s.Bucket)
	}
	u, err := url.Parse(*s.Bucket)
	if err != nil {
		return "", "", "", err
	}
	switch u.Scheme {

	case string(s3Bucket):
		var key string
		if u.Path != "" {
			key = strings.Trim(u.Path, "/")
		}
		return s3Bucket, u.Host, key, nil

	case string(minioBucket):
		var key string
		if u.Path != "" {
			key = strings.Trim(u.Path, "/")
		}
		return minioBucket, u.Host, key, nil

	case "":
		tokens := strings.SplitAfterN(strings.Trim(u.Path, "/"), "/", 2)
		if len(tokens) == 0 {
			return "", "", "", fmt.Errorf("invalid S3 bucket URI: %q", u.String())
		}
		key, bucket := "", strings.Trim(tokens[0], "/")
		if len(tokens) > 1 {
			key = tokens[1]
		}
		return s3Bucket, bucket, key, nil

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
