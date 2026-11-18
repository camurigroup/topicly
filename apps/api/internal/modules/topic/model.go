package topic

import "time"

// Author is a compact user summary embedded in topics and replies.
type Author struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
}

// Topic is a discussion thread.
type Topic struct {
	ID            string    `json:"id"`
	CategoryID    string    `json:"categoryId"`
	CategorySlug  string    `json:"categorySlug"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	Content       string    `json:"content"`
	Author        Author    `json:"author"`
	IsPinned      bool      `json:"isPinned"`
	IsLocked      bool      `json:"isLocked"`
	IsFeatured    bool      `json:"isFeatured"`
	Status        string    `json:"status"`
	ViewCount     int       `json:"viewCount"`
	ReplyCount    int       `json:"replyCount"`
	LikeCount     int       `json:"likeCount"`
	FollowerCount int       `json:"followerCount"`
	Liked         bool      `json:"liked"`
	Following     bool      `json:"following"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	LastActivity  time.Time `json:"lastActivityAt"`
}

// Reply is a response to a topic (optionally nested under another reply).
type Reply struct {
	ID        string     `json:"id"`
	TopicID   string     `json:"topicId"`
	ParentID  *string    `json:"parentId"`
	Content   string     `json:"content"`
	Author    Author     `json:"author"`
	LikeCount int        `json:"likeCount"`
	Liked     bool       `json:"liked"`
	IsDeleted bool       `json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	EditedAt  *time.Time `json:"editedAt"`
}

// CreateTopicInput is the payload to open a topic.
type CreateTopicInput struct {
	CategoryID string `json:"categoryId"`
	Title      string `json:"title"`
	Content    string `json:"content"`
}

// UpdateTopicInput is the payload to edit a topic.
type UpdateTopicInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// CreateReplyInput is the payload to post a reply.
type CreateReplyInput struct {
	Content  string  `json:"content"`
	ParentID *string `json:"parentId"`
}

// UpdateReplyInput is the payload to edit a reply.
type UpdateReplyInput struct {
	Content string `json:"content"`
}

// Like/follow target types.
const (
	TargetTopic = "topic"
	TargetReply = "reply"
	FollowTopic = "topic"
)
