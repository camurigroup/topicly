package storage

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/forum-framework/api/internal/config"
)

type localProvider struct {
	basePath  string
	publicURL string
}

func newLocal(cfg config.StorageConfig) (*localProvider, error) {
	base := cfg.LocalPath
	if base == "" {
		base = "./storage"
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		return nil, err
	}
	return &localProvider{
		basePath:  base,
		publicURL: strings.TrimRight(cfg.PublicURL, "/"),
	}, nil
}

func (p *localProvider) Save(_ context.Context, key string, r io.Reader, _ int64, _ string) (string, error) {
	dest := filepath.Join(p.basePath, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	f, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return p.PublicURL(key), nil
}

func (p *localProvider) Delete(_ context.Context, key string) error {
	dest := filepath.Join(p.basePath, filepath.FromSlash(key))
	err := os.Remove(dest)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (p *localProvider) PublicURL(key string) string {
	return p.publicURL + "/" + strings.TrimLeft(key, "/")
}

// FileServer serves stored files over HTTP (used when provider is local).
func (p *localProvider) FileServer() http.Handler {
	return http.StripPrefix("/files/", http.FileServer(http.Dir(p.basePath)))
}
