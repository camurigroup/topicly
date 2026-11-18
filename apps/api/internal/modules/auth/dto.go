package auth

import "github.com/forum-framework/api/internal/modules/user"

// RegisterInput is the payload for account creation.
type RegisterInput struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

// LoginInput is the payload for authentication (identifier = email or username).
type LoginInput struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// TokenPair is the issued credential set.
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// AuthResult bundles the account view, its roles/permissions and tokens.
type AuthResult struct {
	User        user.Account `json:"user"`
	Roles       []string     `json:"roles"`
	Permissions []string     `json:"permissions"`
	Tokens      TokenPair    `json:"tokens"`
}
