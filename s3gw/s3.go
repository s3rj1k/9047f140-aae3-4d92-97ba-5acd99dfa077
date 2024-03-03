package s3gw

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

const (
	S3DefaultBucketNameEnvKey = "S3_DEFAULT_BUCKET_NAME"
)

func CheckS3BackendLiveliness(ctx context.Context, client *minio.Client) error {
	_, err := client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("error listing buckets: %w", err)
	}

	return nil
}

func MustCreateNewS3BackendConfig(endpoint string, opts *minio.Options) *minio.Client {
	c, err := minio.New(endpoint, opts)
	if err != nil {
		panic(err)
	}

	return c
}

func EnsureBucketExists(ctx context.Context, client *minio.Client, bucketName string) error {
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

func ListObjectsInBucket(ctx context.Context, client *minio.Client, bucketName string) ([]string, error) {
	out := make([]string, 0)

	objectCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{})
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		out = append(out, object.Key)
	}

	return out, nil
}
