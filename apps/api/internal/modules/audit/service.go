package audit

import (
	"context"
	"log/slog"

	"github.com/forum-framework/api/internal/shared/reqmeta"
)

// Service records administrative actions. Recording never fails the caller's
// operation: errors are logged.
type Service struct {
	repo   *Repository
	logger *slog.Logger
}

// NewService builds an audit Service.
func NewService(repo *Repository, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// Record stores an action performed by actorID on a target. The client IP is
// read from the context (see reqmeta). A nil Service is a no-op.
func (s *Service) Record(ctx context.Context, actorID, action, targetType, targetID string, details map[string]any) {
	if s == nil {
		return
	}
	if err := s.repo.Insert(ctx, actorID, action, targetType, targetID, details, reqmeta.IP(ctx)); err != nil {
		s.logger.Error("audit record failed", "action", action, "error", err)
	}
}

// List returns a page of audit entries.
func (s *Service) List(ctx context.Context, f Filter) ([]Entry, int64, error) {
	return s.repo.List(ctx, f)
}

// Actions returns the distinct action names recorded so far.
func (s *Service) Actions(ctx context.Context) ([]string, error) {
	return s.repo.Actions(ctx)
}
