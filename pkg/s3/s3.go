package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"jamel/pkg/fs"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3 struct {
	client *minio.Client
	bucket string
}

func New(
	connect string,
	username, password string,
	_bucket string,
) (*S3, error) {
	_client, err := minio.New(connect, &minio.Options{
		Creds:  credentials.NewStaticV4(username, password, ""),
		Secure: false,
		// Secure: true,
		// Transport: &http.Transport{
		// 	TLSClientConfig: &tls.Config{
		// 		InsecureSkipVerify: true,
		// 	},
		// },
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to s3: %w", err)
	}
	var ctx = context.Background()
	if err := _client.MakeBucket(ctx, _bucket, minio.MakeBucketOptions{}); err != nil {
		exists, errBucketExists := _client.BucketExists(ctx, _bucket)
		if !(errBucketExists == nil && exists) {
			return nil, fmt.Errorf("failed to create bucker: %w", err)
		}
	}
	return &S3{
		client: _client,
		bucket: _bucket,
	}, nil
}

func (s *S3) Upload(filename string) (string, error) {
	file, info, err := fs.OpenFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	var id = uuid.NewString()
	if _, err := s.client.PutObject(context.Background(), s.bucket, id, file, info.Size(), minio.PutObjectOptions{
		ContentType: "text/plain",
	}); err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	return id, nil
}

func (s *S3) Download(id string) (string, error) {
	object, err := s.client.GetObject(context.Background(), s.bucket, id, minio.GetObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve file: %w", err)
	}
	defer object.Close()
	var (
		buf bytes.Buffer
	)
	if _, err := io.Copy(&buf, object); err != nil {
		return "", fmt.Errorf("failed to write obj to buf: %w", err)
	}
	if err := fs.WriteFile(id, buf.Bytes()); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return id, nil
}
