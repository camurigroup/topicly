package user

import "time"

// User is the full user record (includes sensitive fields).
type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	DisplayName  string
	Bio          *string
	AvatarURL    *string
	BannerURL    *string
	Reputation   int
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastActiveAt *time.Time
}

// Status values.
const (
	StatusActive   = "active"
	StatusBanned   = "banned"
	StatusDisabled = "disabled"
)

// PublicProfile is the safe, public-facing view of a user.
type PublicProfile struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	DisplayName  string     `json:"displayName"`
	Bio          *string    `json:"bio"`
	AvatarURL    *string    `json:"avatarUrl"`
	BannerURL    *string    `json:"bannerUrl"`
	Reputation   int        `json:"reputation"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	LastActiveAt *time.Time `json:"lastActiveAt"`
}

// Account is the private view returned to the authenticated owner.
type Account struct {
	PublicProfile
	Email string `json:"email"`
}

// ToPublic maps a User to its public profile.
func (u *User) ToPublic() PublicProfile {
	return PublicProfile{
		ID:           u.ID,
		Username:     u.Username,
		DisplayName:  u.DisplayName,
		Bio:          u.Bio,
		AvatarURL:    u.AvatarURL,
		BannerURL:    u.BannerURL,
		Reputation:   u.Reputation,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt,
		LastActiveAt: u.LastActiveAt,
	}
}

// ToAccount maps a User to its private account view.
func (u *User) ToAccount() Account {
	return Account{PublicProfile: u.ToPublic(), Email: u.Email}
}
