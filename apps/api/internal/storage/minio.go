package storage

import (
	"context"
	"io"
	"strings"

	"github.com/forum-framework/api/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioProvider struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

func newMinio(cfg config.StorageConfig) (*minioProvider, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}

	// Best-effort bucket bootstrap.
	ctx := context.Background()
	if exists, err := client.BucketExists(ctx, cfg.MinioBucket); err == nil && !exists {
		_ = client.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{})
	}

	return &minioProvider{
		client:    client,
		bucket:    cfg.MinioBucket,
		publicURL: strings.TrimRight(cfg.PublicURL, "/"),
	}, nil
}

func (p *minioProvider) Save(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	if _, err := p.client.PutObject(ctx, p.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return "", err
	}
	return p.PublicURL(key), nil
}

func (p *minioProvider) Delete(ctx context.Context, key string) error {
	return p.client.RemoveObject(ctx, p.bucket, key, minio.RemoveObjectOptions{})
}

func (p *minioProvider) PublicURL(key string) string {
	return p.publicURL + "/" + strings.TrimLeft(key, "/")
}
