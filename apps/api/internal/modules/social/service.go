package social

import (
	"context"

	"github.com/forum-framework/api/internal/modules/notification"
	"github.com/forum-framework/api/internal/shared/apperr"
)

// Service implements following users and categories.
type Service struct {
	repo   *Repository
	notify *notification.Service
}

// NewService builds a social Service. notifySvc may be nil.
func NewService(repo *Repository, notifySvc *notification.Service) *Service {
	return &Service{repo: repo, notify: notifySvc}
}

// ToggleUser follows or unfollows a user; a new follow notifies the followed user.
func (s *Service) ToggleUser(ctx context.Context, actorID, actorUsername, username string) (*ToggleResult, error) {
	targetID, err := s.repo.UserIDByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if targetID == actorID {
		return nil, validationErr("você não pode seguir a si mesmo")
	}
	following, err := s.repo.Toggle(ctx, actorID, TargetUser, targetID)
	if err != nil {
		return nil, err
	}
	count, err := s.repo.CountFollowers(ctx, TargetUser, targetID)
	if err != nil {
		return nil, err
	}
	if following && s.notify != nil {
		s.notify.Deliver(ctx, notification.Delivery{
			Recipients: []string{targetID},
			Event: notification.Event{
				Type: notification.TypeFollow, ActorID: actorID,
				EntityType: TargetUser, EntityID: actorID,
				Message: "começou a seguir você", Link: "/u/" + actorUsername,
			},
		})
	}
	return &ToggleResult{Following: following, Count: count}, nil
}

// ToggleCategory follows or unfollows a category or subcategory.
func (s *Service) ToggleCategory(ctx context.Context, actorID, slug string) (*ToggleResult, error) {
	categoryID, err := s.repo.CategoryIDBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	following, err := s.repo.Toggle(ctx, actorID, TargetCategory, categoryID)
	if err != nil {
		return nil, err
	}
	count, err := s.repo.CountFollowers(ctx, TargetCategory, categoryID)
	if err != nil {
		return nil, err
	}
	return &ToggleResult{Following: following, Count: count}, nil
}

// UserStats returns profile counters, including the viewer's follow state.
func (s *Service) UserStats(ctx context.Context, username, viewerID string) (*UserStats, error) {
	userID, err := s.repo.UserIDByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	st := &UserStats{IsSelf: viewerID == userID}
	steps := []struct {
		dst *int64
		fn  func() (int64, error)
	}{
		{&st.Followers, func() (int64, error) { return s.repo.CountFollowers(ctx, TargetUser, userID) }},
		{&st.Following, func() (int64, error) { return s.repo.CountFollowing(ctx, userID, TargetUser) }},
		{&st.FollowedCategories, func() (int64, error) { return s.repo.CountFollowing(ctx, userID, TargetCategory) }},
		{&st.Topics, func() (int64, error) { return s.repo.CountTopicsByAuthor(ctx, userID) }},
		{&st.Replies, func() (int64, error) { return s.repo.CountRepliesByAuthor(ctx, userID) }},
	}
	for _, step := range steps {
		if *step.dst, err = step.fn(); err != nil {
			return nil, err
		}
	}
	if !st.IsSelf {
		if st.IsFollowing, err = s.repo.IsFollowing(ctx, viewerID, TargetUser, userID); err != nil {
			return nil, err
		}
	}
	return st, nil
}

// CategoryStats returns category counters and the viewer's follow state.
func (s *Service) CategoryStats(ctx context.Context, slug, viewerID string) (*CategoryStats, error) {
	categoryID, err := s.repo.CategoryIDBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	st := &CategoryStats{}
	if st.Followers, err = s.repo.CountFollowers(ctx, TargetCategory, categoryID); err != nil {
		return nil, err
	}
	if st.Topics, err = s.repo.CountTopicsInCategory(ctx, categoryID); err != nil {
		return nil, err
	}
	if st.IsFollowing, err = s.repo.IsFollowing(ctx, viewerID, TargetCategory, categoryID); err != nil {
		return nil, err
	}
	return st, nil
}

// Followers lists a user's followers with the total.
func (s *Service) Followers(ctx context.Context, username string, limit, offset int) ([]UserSummary, int64, error) {
	userID, err := s.repo.UserIDByUsername(ctx, username)
	if err != nil {
		return nil, 0, err
	}
	items, err := s.repo.Followers(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountFollowers(ctx, TargetUser, userID)
	return items, total, err
}

// Following lists who a user follows with the total.
func (s *Service) Following(ctx context.Context, username string, limit, offset int) ([]UserSummary, int64, error) {
	userID, err := s.repo.UserIDByUsername(ctx, username)
	if err != nil {
		return nil, 0, err
	}
	items, err := s.repo.Following(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountFollowing(ctx, userID, TargetUser)
	return items, total, err
}

func validationErr(msg string) error { return &validationError{msg: msg} }

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }
func (e *validationError) Unwrap() error { return apperr.ErrValidation }
