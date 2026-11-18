package social

// Follow target types (stored in follows.target_type).
const (
	TargetUser     = "user"
	TargetCategory = "category"
)

// UserSummary is the public card of a user in follower/following lists.
type UserSummary struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
	Bio         *string `json:"bio"`
	Reputation  int     `json:"reputation"`
}

// UserStats are the social counters shown on a profile.
type UserStats struct {
	Followers          int64 `json:"followers"`
	Following          int64 `json:"following"`
	FollowedCategories int64 `json:"followedCategories"`
	Topics             int64 `json:"topics"`
	Replies            int64 `json:"replies"`
	IsFollowing        bool  `json:"isFollowing"`
	IsSelf             bool  `json:"isSelf"`
}

// CategoryStats are the counters shown on a category page.
type CategoryStats struct {
	Followers   int64 `json:"followers"`
	Topics      int64 `json:"topics"`
	IsFollowing bool  `json:"isFollowing"`
}

// ToggleResult is returned by follow toggles.
type ToggleResult struct {
	Following bool  `json:"following"`
	Count     int64 `json:"count"`
}
