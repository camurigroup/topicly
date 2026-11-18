package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/shared/apperr"
)

const userColumns = "id, username, email, password_hash, display_name, bio, avatar_url, banner_url, reputation, status, created_at, updated_at, last_active_at"

// Repository persists users.
type Repository struct {
	store *database.Store
}

// NewRepository builds a user Repository.
func NewRepository(store *database.Store) *Repository {
	return &Repository{store: store}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (*User, error) {
	var (
		u          User
		bio        sql.NullString
		avatar     sql.NullString
		banner     sql.NullString
		lastActive sql.NullTime
	)
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.DisplayName,
		&bio, &avatar, &banner, &u.Reputation, &u.Status,
		&u.CreatedAt, &u.UpdatedAt, &lastActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if bio.Valid {
		u.Bio = &bio.String
	}
	if avatar.Valid {
		u.AvatarURL = &avatar.String
	}
	if banner.Valid {
		u.BannerURL = &banner.String
	}
	if lastActive.Valid {
		t := lastActive.Time
		u.LastActiveAt = &t
	}
	return &u, nil
}

// Create inserts a new user.
func (r *Repository) Create(ctx context.Context, u *User) error {
	q := r.store.Rebind(`INSERT INTO users
        (id, username, email, password_hash, display_name, bio, avatar_url, banner_url, reputation, status, created_at, updated_at, last_active_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	_, err := r.store.DB.ExecContext(ctx, q,
		u.ID, u.Username, u.Email, u.PasswordHash, u.DisplayName,
		u.Bio, u.AvatarURL, u.BannerURL, u.Reputation, u.Status,
		u.CreatedAt, u.UpdatedAt, u.LastActiveAt,
	)
	return err
}

// Count returns the total number of users.
func (r *Repository) Count(ctx context.Context) (int, error) {
	var n int
	err := r.store.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&n)
	return n, err
}

func (r *Repository) getBy(ctx context.Context, column, value string) (*User, error) {
	q := r.store.Rebind("SELECT " + userColumns + " FROM users WHERE " + column + " = ?")
	return scanUser(r.store.DB.QueryRowContext(ctx, q, value))
}

// GetByID fetches a user by id.
func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	return r.getBy(ctx, "id", id)
}

// GetByUsername fetches a user by username.
func (r *Repository) GetByUsername(ctx context.Context, username string) (*User, error) {
	return r.getBy(ctx, "username", username)
}

// GetByEmail fetches a user by email.
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	return r.getBy(ctx, "email", email)
}

func (r *Repository) exists(ctx context.Context, column, value string) (bool, error) {
	q := r.store.Rebind("SELECT 1 FROM users WHERE " + column + " = ? LIMIT 1")
	var one int
	err := r.store.DB.QueryRowContext(ctx, q, value).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// ExistsByUsername reports whether a username is taken.
func (r *Repository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return r.exists(ctx, "username", username)
}

// ExistsByEmail reports whether an email is taken.
func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.exists(ctx, "email", email)
}

// UpdateProfile updates editable profile fields.
func (r *Repository) UpdateProfile(ctx context.Context, id, displayName string, bio *string) error {
	q := r.store.Rebind(`UPDATE users SET display_name = ?, bio = ?, updated_at = ? WHERE id = ?`)
	res, err := r.store.DB.ExecContext(ctx, q, displayName, bio, time.Now().UTC(), id)
	return affected(res, err)
}

// UpdatePassword sets a new password hash.
func (r *Repository) UpdatePassword(ctx context.Context, id, hash string) error {
	q := r.store.Rebind(`UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`)
	res, err := r.store.DB.ExecContext(ctx, q, hash, time.Now().UTC(), id)
	return affected(res, err)
}

// UpdateEmail sets a new email.
func (r *Repository) UpdateEmail(ctx context.Context, id, email string) error {
	q := r.store.Rebind(`UPDATE users SET email = ?, updated_at = ? WHERE id = ?`)
	res, err := r.store.DB.ExecContext(ctx, q, email, time.Now().UTC(), id)
	return affected(res, err)
}

// UpdateLastActive stamps the user's last activity time.
func (r *Repository) UpdateLastActive(ctx context.Context, id string, t time.Time) error {
	q := r.store.Rebind(`UPDATE users SET last_active_at = ? WHERE id = ?`)
	_, err := r.store.DB.ExecContext(ctx, q, t, id)
	return err
}

func affected(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

// UpdateAvatar sets the user's avatar URL.
func (r *Repository) UpdateAvatar(ctx context.Context, id string, url *string) error {
	q := r.store.Rebind(`UPDATE users SET avatar_url = ?, updated_at = ? WHERE id = ?`)
	res, err := r.store.DB.ExecContext(ctx, q, url, time.Now().UTC(), id)
	return affected(res, err)
}

// UpdateBanner sets the user's banner URL.
func (r *Repository) UpdateBanner(ctx context.Context, id string, url *string) error {
	q := r.store.Rebind(`UPDATE users SET banner_url = ?, updated_at = ? WHERE id = ?`)
	res, err := r.store.DB.ExecContext(ctx, q, url, time.Now().UTC(), id)
	return affected(res, err)
}
