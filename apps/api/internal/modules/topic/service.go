package topic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/modules/notification"
	"github.com/forum-framework/api/internal/modules/rbac"
	"github.com/forum-framework/api/internal/shared/apperr"
	"github.com/forum-framework/api/internal/shared/id"
	"github.com/forum-framework/api/internal/shared/slug"
)

// Service implements topic, reply, like and follow business logic.
type Service struct {
	repo   *Repository
	rbac   *rbac.Service
	notify *notification.Service
}

// NewService builds a topic Service. notifySvc may be nil (no notifications).
func NewService(repo *Repository, rbacSvc *rbac.Service, notifySvc *notification.Service) *Service {
	return &Service{repo: repo, rbac: rbacSvc, notify: notifySvc}
}

const (
	maxTitle   = 200
	maxContent = 40000
)

// CreateTopic validates and opens a new topic.
func (s *Service) CreateTopic(ctx context.Context, authorID string, in CreateTopicInput) (*Topic, error) {
	title := strings.TrimSpace(in.Title)
	content := strings.TrimSpace(in.Content)
	if l := len([]rune(title)); l < 3 || l > maxTitle {
		return nil, validationErr("o título deve ter entre 3 e 200 caracteres")
	}
	if content == "" || len([]rune(content)) > maxContent {
		return nil, validationErr("conteúdo inválido")
	}
	if exists, err := s.repo.CategoryExists(ctx, in.CategoryID); err != nil {
		return nil, err
	} else if !exists {
		return nil, validationErr("categoria inexistente")
	}

	topicID := id.New()
	base := slug.Make(title)
	if base == "" {
		base = "topico"
	}
	now := time.Now().UTC()
	t := &Topic{
		ID:           topicID,
		CategoryID:   in.CategoryID,
		Title:        title,
		Slug:         base + "-" + topicID[:8],
		Content:      content,
		Status:       "published",
		CreatedAt:    now,
		UpdatedAt:    now,
		LastActivity: now,
	}
	if err := s.repo.CreateTopic(ctx, t, authorID); err != nil {
		return nil, err
	}
	created, err := s.repo.GetTopicByID(ctx, topicID)
	if err != nil {
		return nil, err
	}
	s.notifyNewTopic(ctx, created)
	return created, nil
}

// GetTopic returns a topic, increments its view count and resolves viewer flags.
func (s *Service) GetTopic(ctx context.Context, topicID, viewerID string) (*Topic, error) {
	t, err := s.repo.GetTopicByID(ctx, topicID)
	if err != nil {
		return nil, err
	}
	_ = s.repo.IncrementView(ctx, topicID)
	t.ViewCount++
	if viewerID != "" {
		t.Liked, _ = s.repo.IsLiked(ctx, viewerID, TargetTopic, topicID)
		t.Following, _ = s.repo.IsFollowingTopic(ctx, viewerID, topicID)
	}
	return t, nil
}

// ListTopics returns a page of topics, optionally filtered by category slug.
func (s *Service) ListTopics(ctx context.Context, categorySlug string, limit, offset int) ([]Topic, int64, error) {
	categoryID := ""
	if categorySlug != "" {
		cid, err := s.repo.CategoryIDBySlug(ctx, categorySlug)
		if err != nil {
			if errors.Is(err, apperr.ErrNotFound) {
				return []Topic{}, 0, nil
			}
			return nil, 0, err
		}
		categoryID = cid
	}
	topics, err := s.repo.ListTopics(ctx, categoryID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountTopics(ctx, categoryID)
	if err != nil {
		return nil, 0, err
	}
	return topics, total, nil
}

// UpdateTopic edits a topic if the actor is the author or holds topics.edit.
func (s *Service) UpdateTopic(ctx context.Context, actorID, topicID string, in UpdateTopicInput) (*Topic, error) {
	current, err := s.repo.GetTopicByID(ctx, topicID)
	if err != nil {
		return nil, err
	}
	if current.Author.ID != actorID && !s.can(ctx, actorID, rbac.PermTopicsEdit) {
		return nil, apperr.ErrForbidden
	}
	title := strings.TrimSpace(in.Title)
	content := strings.TrimSpace(in.Content)
	if l := len([]rune(title)); l < 3 || l > maxTitle {
		return nil, validationErr("o título deve ter entre 3 e 200 caracteres")
	}
	if content == "" || len([]rune(content)) > maxContent {
		return nil, validationErr("conteúdo inválido")
	}
	_ = s.repo.AddEdit(ctx, "topic", topicID, actorID, current.Content)
	if err := s.repo.UpdateTopic(ctx, topicID, title, current.Slug, content); err != nil {
		return nil, err
	}
	return s.repo.GetTopicByID(ctx, topicID)
}

// DeleteTopic removes a topic if the actor is the author or holds topics.delete.
func (s *Service) DeleteTopic(ctx context.Context, actorID, topicID string) error {
	author, err := s.repo.TopicAuthor(ctx, topicID)
	if err != nil {
		return err
	}
	if author != actorID && !s.can(ctx, actorID, rbac.PermTopicsDelete) {
		return apperr.ErrForbidden
	}
	return s.repo.DeleteTopic(ctx, topicID)
}

// SetFlag toggles a topic flag (is_pinned, is_locked, is_featured).
func (s *Service) SetFlag(ctx context.Context, topicID, column string, value bool) error {
	return s.repo.SetFlag(ctx, topicID, column, value)
}

// CreateReply posts a reply to an unlocked topic.
func (s *Service) CreateReply(ctx context.Context, authorID, topicID string, in CreateReplyInput) (*Reply, error) {
	t, err := s.repo.GetTopicByID(ctx, topicID)
	if err != nil {
		return nil, err
	}
	if t.IsLocked {
		return nil, apperr.ErrForbidden
	}
	content := strings.TrimSpace(in.Content)
	if content == "" || len([]rune(content)) > maxContent {
		return nil, validationErr("conteúdo inválido")
	}
	parentAuthor := ""
	if in.ParentID != nil {
		pa, parentTopic, err := s.repo.ReplyMeta(ctx, *in.ParentID)
		if err != nil || parentTopic != topicID {
			return nil, validationErr("resposta pai inválida")
		}
		parentAuthor = pa
	}
	replyID := id.New()
	rp := &Reply{ID: replyID, TopicID: topicID, ParentID: in.ParentID, Content: content}
	if err := s.repo.CreateReply(ctx, rp, authorID); err != nil {
		return nil, err
	}
	created, err := s.repo.GetReplyByID(ctx, replyID)
	if err != nil {
		return nil, err
	}
	s.notifyReply(ctx, t, created, parentAuthor)
	return created, nil
}

// ListReplies returns a page of replies with viewer like flags.
func (s *Service) ListReplies(ctx context.Context, topicID, viewerID string, limit, offset int) ([]Reply, int64, error) {
	replies, err := s.repo.ListReplies(ctx, topicID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountReplies(ctx, topicID)
	if err != nil {
		return nil, 0, err
	}
	if viewerID != "" && len(replies) > 0 {
		ids := make([]string, len(replies))
		for i := range replies {
			ids[i] = replies[i].ID
		}
		liked, err := s.repo.LikedSet(ctx, viewerID, TargetReply, ids)
		if err != nil {
			return nil, 0, err
		}
		for i := range replies {
			replies[i].Liked = liked[replies[i].ID]
		}
	}
	return replies, total, nil
}

// UpdateReply edits a reply if the actor is the author or holds replies.edit.
func (s *Service) UpdateReply(ctx context.Context, actorID, replyID string, in UpdateReplyInput) (*Reply, error) {
	current, err := s.repo.GetReplyByID(ctx, replyID)
	if err != nil {
		return nil, err
	}
	if current.Author.ID != actorID && !s.can(ctx, actorID, rbac.PermRepliesEdit) {
		return nil, apperr.ErrForbidden
	}
	content := strings.TrimSpace(in.Content)
	if content == "" || len([]rune(content)) > maxContent {
		return nil, validationErr("conteúdo inválido")
	}
	_ = s.repo.AddEdit(ctx, "reply", replyID, actorID, current.Content)
	if err := s.repo.UpdateReply(ctx, replyID, content); err != nil {
		return nil, err
	}
	return s.repo.GetReplyByID(ctx, replyID)
}

// DeleteReply soft-deletes a reply if the actor is the author or holds replies.delete.
func (s *Service) DeleteReply(ctx context.Context, actorID, replyID string) error {
	author, topicID, err := s.repo.ReplyMeta(ctx, replyID)
	if err != nil {
		return err
	}
	if author != actorID && !s.can(ctx, actorID, rbac.PermRepliesDelete) {
		return apperr.ErrForbidden
	}
	return s.repo.SoftDeleteReply(ctx, replyID, topicID)
}

// ToggleTopicLike likes or unlikes a topic.
func (s *Service) ToggleTopicLike(ctx context.Context, userID, topicID string) (bool, int, error) {
	t, err := s.repo.GetTopicByID(ctx, topicID)
	if err != nil {
		return false, 0, err
	}
	liked, count, err := s.repo.ToggleLike(ctx, userID, TargetTopic, topicID)
	if err == nil && liked && s.notify != nil {
		s.notify.Deliver(ctx, notification.Delivery{
			Recipients: []string{t.Author.ID},
			Event: notification.Event{
				Type: notification.TypeTopicLike, ActorID: userID,
				EntityType: TargetTopic, EntityID: t.ID,
				Message: "curtiu seu tópico " + quote(t.Title), Link: topicLink(t.ID, ""),
			},
		})
	}
	return liked, count, err
}

// ToggleReplyLike likes or unlikes a reply.
func (s *Service) ToggleReplyLike(ctx context.Context, userID, replyID string) (bool, int, error) {
	authorID, topicID, err := s.repo.ReplyMeta(ctx, replyID)
	if err != nil {
		return false, 0, err
	}
	liked, count, err := s.repo.ToggleLike(ctx, userID, TargetReply, replyID)
	if err == nil && liked && s.notify != nil {
		title := ""
		if t, terr := s.repo.GetTopicByID(ctx, topicID); terr == nil {
			title = t.Title
		}
		s.notify.Deliver(ctx, notification.Delivery{
			Recipients: []string{authorID},
			Event: notification.Event{
				Type: notification.TypeReplyLike, ActorID: userID,
				EntityType: TargetReply, EntityID: replyID,
				Message: "curtiu sua resposta em " + quote(title), Link: topicLink(topicID, replyID),
			},
		})
	}
	return liked, count, err
}

// ToggleTopicFollow follows or unfollows a topic.
func (s *Service) ToggleTopicFollow(ctx context.Context, userID, topicID string) (bool, int, error) {
	if _, err := s.repo.TopicAuthor(ctx, topicID); err != nil {
		return false, 0, err
	}
	return s.repo.ToggleFollowTopic(ctx, userID, topicID)
}

func (s *Service) can(ctx context.Context, userID, perm string) bool {
	ok, err := s.rbac.HasPermission(ctx, userID, perm)
	return err == nil && ok
}

func validationErr(msg string) error { return &validationError{msg: msg} }

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }
func (e *validationError) Unwrap() error { return apperr.ErrValidation }

// Discover returns a page of topics for a discovery section
// (recent, popular or trending).
func (s *Service) Discover(ctx context.Context, section string, limit, offset int) ([]Topic, int64, error) {
	switch section {
	case DiscoverRecent, DiscoverPopular, DiscoverTrending:
	default:
		return nil, 0, validationErr("seção inválida (use recent, popular ou trending)")
	}
	return s.repo.Discover(ctx, section, limit, offset)
}
