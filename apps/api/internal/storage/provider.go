// Package storage abstracts file persistence behind a provider interface so
// the app can use local disk or an S3-compatible service (MinIO) interchangeably.
package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/forum-framework/api/internal/config"
)

// Provider persists and locates uploaded files.
type Provider interface {
	// Save stores the object under key and returns its public URL.
	Save(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error)
	// Delete removes the object at key.
	Delete(ctx context.Context, key string) error
	// PublicURL returns the public URL for a stored key.
	PublicURL(key string) string
}

// New builds a Provider from configuration.
func New(cfg config.StorageConfig) (Provider, error) {
	switch strings.ToLower(cfg.Provider) {
	case "", "local":
		return newLocal(cfg)
	case "minio", "s3":
		return newMinio(cfg)
	default:
		return nil, fmt.Errorf("unsupported STORAGE_PROVIDER %q (use local or minio)", cfg.Provider)
	}
}
