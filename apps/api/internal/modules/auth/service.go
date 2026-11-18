package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/modules/rbac"
	"github.com/forum-framework/api/internal/modules/user"
	"github.com/forum-framework/api/internal/shared/apperr"
	"github.com/forum-framework/api/internal/shared/id"
	"github.com/forum-framework/api/internal/shared/security"
	"github.com/forum-framework/api/internal/shared/validate"
)

// Config carries token settings for the auth service.
type Config struct {
	AccessSecret string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
}

// Service implements registration, login and token lifecycle.
type Service struct {
	users  *user.Repository
	tokens *RefreshTokenRepository
	rbac   *rbac.Service
	cfg    Config
}

// NewService builds the auth Service.
func NewService(users *user.Repository, tokens *RefreshTokenRepository, rbacSvc *rbac.Service, cfg Config) *Service {
	return &Service{users: users, tokens: tokens, rbac: rbacSvc, cfg: cfg}
}

// Register validates input, creates the account and issues tokens. The very
// first user created becomes the Owner.
func (s *Service) Register(ctx context.Context, in RegisterInput, userAgent, ip string) (*AuthResult, error) {
	in.Username = strings.TrimSpace(in.Username)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if in.DisplayName == "" {
		in.DisplayName = in.Username
	}

	if err := validate.Username(in.Username); err != nil {
		return nil, wrapValidation(err)
	}
	if err := validate.Email(in.Email); err != nil {
		return nil, wrapValidation(err)
	}
	if err := validate.Password(in.Password); err != nil {
		return nil, wrapValidation(err)
	}
	if err := validate.DisplayName(in.DisplayName); err != nil {
		return nil, wrapValidation(err)
	}

	if taken, err := s.users.ExistsByUsername(ctx, in.Username); err != nil {
		return nil, err
	} else if taken {
		return nil, fieldConflict("username", "username já está em uso")
	}
	if taken, err := s.users.ExistsByEmail(ctx, in.Email); err != nil {
		return nil, err
	} else if taken {
		return nil, fieldConflict("email", "e-mail já está em uso")
	}

	hash, err := security.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	count, err := s.users.Count(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	u := &user.User{
		ID:           id.New(),
		Username:     in.Username,
		Email:        in.Email,
		PasswordHash: hash,
		DisplayName:  in.DisplayName,
		Reputation:   0,
		Status:       user.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}

	if count == 0 {
		if err := s.rbac.AssignRoleBySlug(ctx, u.ID, rbac.RoleOwner); err != nil {
			return nil, err
		}
	}

	return s.buildResult(ctx, u, userAgent, ip)
}

// Login authenticates by email or username and issues tokens.
func (s *Service) Login(ctx context.Context, in LoginInput, userAgent, ip string) (*AuthResult, error) {
	identifier := strings.TrimSpace(in.Identifier)
	if identifier == "" || in.Password == "" {
		return nil, apperr.ErrUnauthorized
	}

	var (
		u   *user.User
		err error
	)
	if strings.Contains(identifier, "@") {
		u, err = s.users.GetByEmail(ctx, strings.ToLower(identifier))
	} else {
		u, err = s.users.GetByUsername(ctx, identifier)
	}
	if errors.Is(err, apperr.ErrNotFound) {
		return nil, apperr.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	ok, err := security.VerifyPassword(in.Password, u.PasswordHash)
	if err != nil || !ok {
		return nil, apperr.ErrUnauthorized
	}
	if u.Status == user.StatusBanned || u.Status == user.StatusDisabled {
		return nil, apperr.ErrForbidden
	}

	_ = s.users.UpdateLastActive(ctx, u.ID, time.Now().UTC())
	return s.buildResult(ctx, u, userAgent, ip)
}

// Refresh rotates a refresh token and issues a new token pair.
func (s *Service) Refresh(ctx context.Context, rawRefresh, userAgent, ip string) (*TokenPair, error) {
	rawRefresh = strings.TrimSpace(rawRefresh)
	if rawRefresh == "" {
		return nil, apperr.ErrUnauthorized
	}
	hash := security.HashRefreshToken(rawRefresh)
	current, err := s.tokens.FindActiveByHash(ctx, hash)
	if errors.Is(err, apperr.ErrNotFound) {
		return nil, apperr.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	pair, newID, err := s.issueTokens(ctx, current.UserID, userAgent, ip)
	if err != nil {
		return nil, err
	}
	if err := s.tokens.Revoke(ctx, current.ID, &newID); err != nil {
		return nil, err
	}
	return pair, nil
}

// Logout revokes a single refresh token.
func (s *Service) Logout(ctx context.Context, rawRefresh string) error {
	rawRefresh = strings.TrimSpace(rawRefresh)
	if rawRefresh == "" {
		return nil
	}
	return s.tokens.RevokeByHash(ctx, security.HashRefreshToken(rawRefresh))
}

func (s *Service) buildResult(ctx context.Context, u *user.User, userAgent, ip string) (*AuthResult, error) {
	pair, _, err := s.issueTokens(ctx, u.ID, userAgent, ip)
	if err != nil {
		return nil, err
	}
	roles, err := s.rbac.RolesForUser(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	perms, err := s.rbac.PermissionsForUser(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	return &AuthResult{
		User:        u.ToAccount(),
		Roles:       roles,
		Permissions: perms,
		Tokens:      *pair,
	}, nil
}

func (s *Service) issueTokens(ctx context.Context, userID, userAgent, ip string) (*TokenPair, string, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	access, err := security.GenerateAccessToken(s.cfg.AccessSecret, u.ID, u.Username, s.cfg.AccessTTL)
	if err != nil {
		return nil, "", err
	}
	raw, hash, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, "", err
	}
	rtID := id.New()
	now := time.Now().UTC()
	rt := &RefreshToken{
		ID:        rtID,
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: now.Add(s.cfg.RefreshTTL),
		CreatedAt: now,
	}
	if err := s.tokens.Create(ctx, rt, userAgent, ip); err != nil {
		return nil, "", err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int64(s.cfg.AccessTTL.Seconds()),
	}, rtID, nil
}
