package admin

import "time"

// Totals are the dashboard counters, all read from the database.
type Totals struct {
	Users          int64 `json:"users"`
	ActiveUsers    int64 `json:"activeUsers"` // active in the last 7 days
	NewUsers       int64 `json:"newUsers"`    // registered in the last 30 days
	Categories     int64 `json:"categories"`
	Subcategories  int64 `json:"subcategories"`
	Topics         int64 `json:"topics"`
	Replies        int64 `json:"replies"`
	Likes          int64 `json:"likes"`
	ReportsPending int64 `json:"reportsPending"`
	Reports        int64 `json:"reports"`
	Banned         int64 `json:"banned"`
	Muted          int64 `json:"muted"`
	Warnings       int64 `json:"warnings"`
}

// Series metrics (allow-listed tables).
var seriesTables = map[string]string{
	"users":   "users",
	"topics":  "topics",
	"replies": "replies",
	"likes":   "likes",
	"reports": "reports",
}

// UserRow is one line of the admin users list.
type UserRow struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	DisplayName  string     `json:"displayName"`
	Email        string     `json:"email"`
	AvatarURL    *string    `json:"avatarUrl"`
	Reputation   int        `json:"reputation"`
	Roles        []string   `json:"roles"`
	Topics       int64      `json:"topics"`
	Replies      int64      `json:"replies"`
	IsBanned     bool       `json:"isBanned"`
	IsMuted      bool       `json:"isMuted"`
	CreatedAt    time.Time  `json:"createdAt"`
	LastActiveAt *time.Time `json:"lastActiveAt"`
}

// UserFilter narrows the users list.
type UserFilter struct {
	Q      string
	Role   string // role slug, or "member" for users without roles
	Status string // active | banned | muted
	Sort   string // created | active | username | reputation | topics
	Order  string
	Limit  int
	Offset int
}

// TopicRow is one line of the admin topics list.
type TopicRow struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	CategorySlug string    `json:"categorySlug"`
	CategoryName string    `json:"categoryName"`
	Author       string    `json:"author"`
	IsPinned     bool      `json:"isPinned"`
	IsLocked     bool      `json:"isLocked"`
	IsFeatured   bool      `json:"isFeatured"`
	Views        int64     `json:"views"`
	Replies      int64     `json:"replies"`
	Likes        int64     `json:"likes"`
	CreatedAt    time.Time `json:"createdAt"`
	LastActivity time.Time `json:"lastActivityAt"`
}

// TopicFilter narrows the topics list.
type TopicFilter struct {
	Q        string
	Category string // category slug
	Flag     string // pinned | locked | featured
	Sort     string // created | activity | replies | likes | views | title
	Order    string
	Limit    int
	Offset   int
}

// ReplyRow is one line of the admin replies list.
type ReplyRow struct {
	ID         string    `json:"id"`
	Excerpt    string    `json:"excerpt"`
	TopicID    string    `json:"topicId"`
	TopicTitle string    `json:"topicTitle"`
	Author     string    `json:"author"`
	Likes      int64     `json:"likes"`
	IsDeleted  bool      `json:"isDeleted"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ReplyFilter narrows the replies list.
type ReplyFilter struct {
	Q      string
	State  string // visible | deleted
	Sort   string // created | likes
	Order  string
	Limit  int
	Offset int
}

// WarningRow is one line of the warnings list.
type WarningRow struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Display   string    `json:"displayName"`
	Reason    string    `json:"reason"`
	IssuedBy  *string   `json:"issuedBy"`
	CreatedAt time.Time `json:"createdAt"`
}

// WarningFilter narrows the warnings list.
type WarningFilter struct {
	Q      string
	Order  string
	Limit  int
	Offset int
}

// StorageInfo describes the active storage provider (secrets redacted).
type StorageInfo struct {
	Provider      string `json:"provider"`
	LocalPath     string `json:"localPath,omitempty"`
	PublicURL     string `json:"publicUrl,omitempty"`
	Endpoint      string `json:"endpoint,omitempty"`
	Bucket        string `json:"bucket,omitempty"`
	UseSSL        bool   `json:"useSsl"`
	MaxImageBytes int64  `json:"maxImageBytes"`
	Avatars       int64  `json:"avatars"`
	Banners       int64  `json:"banners"`
}
