package notification

import "time"

// Notification types.
const (
	TypeFollow        = "follow"         // someone started following you
	TypeTopicLike     = "topic_like"     // someone liked your topic
	TypeReplyLike     = "reply_like"     // someone liked your reply
	TypeReply         = "reply"          // reply on your topic or to your reply
	TypeFollowedReply = "followed_reply" // reply on a topic you follow
	TypeMention       = "mention"        // someone @mentioned you
	TypeNewTopic      = "new_topic"      // new topic by a user/category you follow
	TypePunishment    = "punishment"     // moderation action against you (Etapa 5)
	TypeAdmin         = "admin"          // notice broadcast by an administrator
)

// Actor is the compact user summary of who triggered a notification.
type Actor struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
}

// Notification is an in-app notification delivered to a single user.
type Notification struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Actor      *Actor    `json:"actor"`
	EntityType *string   `json:"entityType"`
	EntityID   *string   `json:"entityId"`
	Message    string    `json:"message"`
	Link       *string   `json:"link"`
	IsRead     bool      `json:"isRead"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Event describes what triggered a notification; recipients are chosen by the caller.
// Message is the action phrase rendered after the actor's name (e.g. "curtiu seu tópico").
type Event struct {
	Type       string
	ActorID    string
	EntityType string
	EntityID   string
	Message    string
	Link       string
}

// StreamEvent is pushed to live subscribers (SSE).
type StreamEvent struct {
	Kind         string        `json:"kind"` // "notification" | "sync"
	Notification *Notification `json:"notification,omitempty"`
	Unread       int64         `json:"unread"`
}
