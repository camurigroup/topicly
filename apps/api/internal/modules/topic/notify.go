package topic

import (
	"context"

	"github.com/forum-framework/api/internal/modules/notification"
)

// Follow target types shared with the social module.
const (
	followUser     = "user"
	followCategory = "category"
)

const maxQuotedTitle = 80

// quote renders a topic title for notification messages, trimming long ones.
func quote(title string) string {
	r := []rune(title)
	if len(r) > maxQuotedTitle {
		title = string(r[:maxQuotedTitle-1]) + "…"
	}
	return "“" + title + "”"
}

// topicLink builds the web path to a topic, optionally anchored at a reply.
func topicLink(topicID, replyID string) string {
	if replyID == "" {
		return "/t/" + topicID
	}
	return "/t/" + topicID + "#reply-" + replyID
}

// notifyNewTopic tells mentioned users and followers of the author and of the
// category (and its parent) about a freshly published topic.
func (s *Service) notifyNewTopic(ctx context.Context, t *Topic) {
	if s.notify == nil {
		return
	}
	base := notification.Event{ActorID: t.Author.ID, EntityType: TargetTopic, EntityID: t.ID, Link: topicLink(t.ID, "")}

	mention := base
	mention.Type, mention.Message = notification.TypeMention, "mencionou você no tópico "+quote(t.Title)

	byAuthor := base
	byAuthor.Type, byAuthor.Message = notification.TypeNewTopic, "publicou um novo tópico: "+quote(t.Title)

	categoryFollowers := s.notify.Followers(ctx, followCategory, t.CategoryID)
	inCategory := base
	inCategory.Type, inCategory.Message = notification.TypeNewTopic, "abriu o tópico "+quote(t.Title)
	if name, parentID, err := s.repo.CategoryInfo(ctx, t.CategoryID); err == nil {
		inCategory.Message += " em " + name
		if parentID != nil {
			categoryFollowers = append(categoryFollowers, s.notify.Followers(ctx, followCategory, *parentID)...)
		}
	}

	s.notify.Deliver(ctx,
		notification.Delivery{Recipients: s.notify.MentionRecipients(ctx, t.Content), Event: mention},
		notification.Delivery{Recipients: s.notify.Followers(ctx, followUser, t.Author.ID), Event: byAuthor},
		notification.Delivery{Recipients: categoryFollowers, Event: inCategory},
	)
}

// notifyReply informs, in priority order (one notification per user): the
// author of the parent reply, the topic author, @mentioned users and the
// topic's followers.
func (s *Service) notifyReply(ctx context.Context, t *Topic, rp *Reply, parentAuthorID string) {
	if s.notify == nil {
		return
	}
	base := notification.Event{ActorID: rp.Author.ID, EntityType: TargetReply, EntityID: rp.ID, Link: topicLink(t.ID, rp.ID)}

	toParent := base
	toParent.Type, toParent.Message = notification.TypeReply, "respondeu ao seu comentário em "+quote(t.Title)

	toAuthor := base
	toAuthor.Type, toAuthor.Message = notification.TypeReply, "respondeu ao seu tópico "+quote(t.Title)

	mention := base
	mention.Type, mention.Message = notification.TypeMention, "mencionou você em "+quote(t.Title)

	toFollowers := base
	toFollowers.Type, toFollowers.Message = notification.TypeFollowedReply, "respondeu em "+quote(t.Title)+", que você segue"

	s.notify.Deliver(ctx,
		notification.Delivery{Recipients: []string{parentAuthorID}, Event: toParent},
		notification.Delivery{Recipients: []string{t.Author.ID}, Event: toAuthor},
		notification.Delivery{Recipients: s.notify.MentionRecipients(ctx, rp.Content), Event: mention},
		notification.Delivery{Recipients: s.notify.Followers(ctx, FollowTopic, t.ID), Event: toFollowers},
	)
}
