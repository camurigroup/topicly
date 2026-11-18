package notification

import (
	"context"
	"log/slog"
	"strings"
)

// Service creates, lists and streams notifications.
type Service struct {
	repo   *Repository
	hub    *Hub
	logger *slog.Logger
}

// NewService builds a notification Service.
func NewService(repo *Repository, hub *Hub, logger *slog.Logger) *Service {
	return &Service{repo: repo, hub: hub, logger: logger}
}

// Delivery pairs a group of recipients with the event they should receive.
type Delivery struct {
	Recipients []string
	Event      Event
}

// Deliver persists and pushes notifications. Deliveries are applied in order and
// a user receives at most one notification per call (the first matching
// delivery wins); the actor never notifies themself. Failures are logged and
// never propagated, so notifications cannot break the action that caused them.
func (s *Service) Deliver(ctx context.Context, deliveries ...Delivery) {
	seen := make(map[string]struct{})
	for _, d := range deliveries {
		var recipients []string
		for _, uid := range d.Recipients {
			if uid == "" || uid == d.Event.ActorID {
				continue
			}
			if _, dup := seen[uid]; dup {
				continue
			}
			seen[uid] = struct{}{}
			recipients = append(recipients, uid)
		}
		if len(recipients) == 0 {
			continue
		}
		created, err := s.repo.InsertMany(ctx, recipients, d.Event)
		if err != nil {
			s.logger.Error("notification insert failed", "type", d.Event.Type, "error", err)
			continue
		}
		s.publishCreated(ctx, recipients, created, d.Event.ActorID)
	}
}

func (s *Service) publishCreated(ctx context.Context, recipients []string, created []Notification, actorID string) {
	var actor *Actor
	actorLoaded := false
	for i, uid := range recipients {
		if !s.hub.Connected(uid) {
			continue
		}
		if !actorLoaded {
			actor, _ = s.repo.ActorSummary(ctx, actorID)
			actorLoaded = true
		}
		n := created[i]
		n.Actor = actor
		unread, _ := s.repo.Count(ctx, uid, true)
		s.hub.Publish(uid, StreamEvent{Kind: "notification", Notification: &n, Unread: unread})
	}
}

// MentionRecipients resolves the users @mentioned in content.
func (s *Service) MentionRecipients(ctx context.Context, content string) []string {
	names := ExtractMentions(content)
	if len(names) == 0 {
		return nil
	}
	ids, err := s.repo.UserIDsByUsernames(ctx, names)
	if err != nil {
		s.logger.Error("mention lookup failed", "error", err)
	}
	return ids
}

// Followers lists who follows a target; errors are logged and yield no recipients.
func (s *Service) Followers(ctx context.Context, targetType, targetID string) []string {
	ids, err := s.repo.FollowerIDs(ctx, targetType, targetID)
	if err != nil {
		s.logger.Error("follower lookup failed", "target", targetType, "error", err)
	}
	return ids
}

// List returns a page of the user's notifications and the total.
func (s *Service) List(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]Notification, int64, error) {
	items, err := s.repo.List(ctx, userID, unreadOnly, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count(ctx, userID, unreadOnly)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Unread returns the number of unread notifications.
func (s *Service) Unread(ctx context.Context, userID string) (int64, error) {
	return s.repo.Count(ctx, userID, true)
}

// MarkRead marks one notification as read.
func (s *Service) MarkRead(ctx context.Context, userID, id string) error {
	if err := s.repo.MarkRead(ctx, userID, id); err != nil {
		return err
	}
	s.sync(ctx, userID)
	return nil
}

// MarkAllRead marks all notifications as read and returns how many changed.
func (s *Service) MarkAllRead(ctx context.Context, userID string) (int64, error) {
	n, err := s.repo.MarkAllRead(ctx, userID)
	if err != nil {
		return 0, err
	}
	s.sync(ctx, userID)
	return n, nil
}

// Delete removes one notification.
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	if err := s.repo.Delete(ctx, userID, id); err != nil {
		return err
	}
	s.sync(ctx, userID)
	return nil
}

// sync tells the user's other open tabs to refresh their counters.
func (s *Service) sync(ctx context.Context, userID string) {
	if !s.hub.Connected(userID) {
		return
	}
	unread, _ := s.repo.Count(ctx, userID, true)
	s.hub.Publish(userID, StreamEvent{Kind: "sync", Unread: unread})
}

// Broadcast sends an administrator notice to every active user.
func (s *Service) Broadcast(ctx context.Context, actorID, message, link string) (int, error) {
	message = strings.TrimSpace(message)
	if l := len([]rune(message)); l < 3 || l > 500 {
		return 0, validationErr("a mensagem deve ter entre 3 e 500 caracteres")
	}
	link = strings.TrimSpace(link)
	if link != "" && (!strings.HasPrefix(link, "/") || strings.HasPrefix(link, "//") || len(link) > 500) {
		return 0, validationErr("o link deve ser um caminho interno, ex.: /t/…")
	}
	ids, err := s.repo.ActiveUserIDs(ctx)
	if err != nil {
		return 0, err
	}
	ev := Event{Type: TypeAdmin, ActorID: actorID, Message: message, Link: link}
	// Admin notices also reach the sender, so the actor is not filtered out here.
	created, err := s.repo.InsertMany(ctx, ids, ev)
	if err != nil {
		return 0, err
	}
	s.publishCreated(ctx, ids, created, actorID)
	return len(created), nil
}

// Subscribe registers a live stream for the user.
func (s *Service) Subscribe(userID string) (<-chan StreamEvent, func()) {
	return s.hub.Subscribe(userID)
}
