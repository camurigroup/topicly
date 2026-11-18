package admin

import (
	"context"
	"time"

	"github.com/forum-framework/api/internal/config"
	"github.com/forum-framework/api/internal/storage"
)

// Service exposes dashboard metrics and admin listings.
type Service struct {
	repo    *Repository
	storage config.StorageConfig
}

// NewService builds the admin Service.
func NewService(repo *Repository, storageCfg config.StorageConfig) *Service {
	return &Service{repo: repo, storage: storageCfg}
}

// Totals returns the dashboard counters.
func (s *Service) Totals(ctx context.Context) (*Totals, error) {
	return s.repo.Totals(ctx, time.Now().UTC())
}

// Series returns a zero-filled time series of a metric for a period.
func (s *Service) Series(ctx context.Context, metric, period string) ([]Point, error) {
	now := time.Now().UTC()
	start, err := WindowStart(period, now)
	if err != nil {
		return nil, err
	}
	daily, err := s.repo.DailyCounts(ctx, metric, start)
	if err != nil {
		return nil, err
	}
	return BuildSeries(period, daily, now)
}

// ListUsers validates filters and returns a page of users.
func (s *Service) ListUsers(ctx context.Context, f UserFilter) ([]UserRow, int64, error) {
	if err := OneOf("status", f.Status, "active", "banned", "muted"); err != nil {
		return nil, 0, err
	}
	return s.repo.ListUsers(ctx, f, time.Now().UTC())
}

// ListTopics validates filters and returns a page of topics.
func (s *Service) ListTopics(ctx context.Context, f TopicFilter) ([]TopicRow, int64, error) {
	if err := OneOf("flag", f.Flag, "pinned", "locked", "featured"); err != nil {
		return nil, 0, err
	}
	return s.repo.ListTopics(ctx, f)
}

// ListReplies validates filters and returns a page of replies.
func (s *Service) ListReplies(ctx context.Context, f ReplyFilter) ([]ReplyRow, int64, error) {
	if err := OneOf("state", f.State, "visible", "deleted"); err != nil {
		return nil, 0, err
	}
	return s.repo.ListReplies(ctx, f)
}

// ListWarnings returns a page of warnings.
func (s *Service) ListWarnings(ctx context.Context, f WarningFilter) ([]WarningRow, int64, error) {
	return s.repo.ListWarnings(ctx, f)
}

// Storage describes the active storage provider without secrets.
func (s *Service) Storage(ctx context.Context) (*StorageInfo, error) {
	avatars, banners, err := s.repo.MediaCounts(ctx)
	if err != nil {
		return nil, err
	}
	info := &StorageInfo{
		Provider:      s.storage.Provider,
		MaxImageBytes: storage.MaxImageBytes,
		Avatars:       avatars,
		Banners:       banners,
	}
	if info.Provider == "" {
		info.Provider = "local"
	}
	if info.Provider == "local" {
		info.LocalPath, info.PublicURL = s.storage.LocalPath, s.storage.PublicURL
	} else {
		info.Endpoint, info.Bucket, info.UseSSL = s.storage.MinioEndpoint, s.storage.MinioBucket, s.storage.MinioUseSSL
	}
	return info, nil
}
