package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/shared/apperr"
)

// RefreshToken is a stored, hashed refresh credential.
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// RefreshTokenRepository persists refresh tokens.
type RefreshTokenRepository struct {
	store *database.Store
}

// NewRefreshTokenRepository builds a RefreshTokenRepository.
func NewRefreshTokenRepository(store *database.Store) *RefreshTokenRepository {
	return &RefreshTokenRepository{store: store}
}

// Create stores a new refresh token.
func (r *RefreshTokenRepository) Create(ctx context.Context, t *RefreshToken, userAgent, ip string) error {
	q := r.store.Rebind(`INSERT INTO refresh_tokens
        (id, user_id, token_hash, expires_at, revoked_at, replaced_by, user_agent, ip, created_at)
        VALUES (?, ?, ?, ?, NULL, NULL, ?, ?, ?)`)
	_, err := r.store.DB.ExecContext(ctx, q, t.ID, t.UserID, t.TokenHash, t.ExpiresAt, userAgent, ip, t.CreatedAt)
	return err
}

// FindActiveByHash returns a non-revoked, non-expired token by its hash.
func (r *RefreshTokenRepository) FindActiveByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	q := r.store.Rebind(`SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
        FROM refresh_tokens WHERE token_hash = ?`)
	var (
		t       RefreshToken
		revoked sql.NullTime
	)
	err := r.store.DB.QueryRowContext(ctx, q, hash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &revoked, &t.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if revoked.Valid {
		return nil, apperr.ErrUnauthorized
	}
	if time.Now().UTC().After(t.ExpiresAt) {
		return nil, apperr.ErrUnauthorized
	}
	return &t, nil
}

// Revoke marks a token revoked, optionally recording its successor.
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string, replacedBy *string) error {
	q := r.store.Rebind("UPDATE refresh_tokens SET revoked_at = ?, replaced_by = ? WHERE id = ? AND revoked_at IS NULL")
	_, err := r.store.DB.ExecContext(ctx, q, time.Now().UTC(), replacedBy, id)
	return err
}

// RevokeByHash revokes a token identified by its hash.
func (r *RefreshTokenRepository) RevokeByHash(ctx context.Context, hash string) error {
	q := r.store.Rebind("UPDATE refresh_tokens SET revoked_at = ? WHERE token_hash = ? AND revoked_at IS NULL")
	_, err := r.store.DB.ExecContext(ctx, q, time.Now().UTC(), hash)
	return err
}

// RevokeAllForUser revokes every active token for a user.
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	q := r.store.Rebind("UPDATE refresh_tokens SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL")
	_, err := r.store.DB.ExecContext(ctx, q, time.Now().UTC(), userID)
	return err
}
