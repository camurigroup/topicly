package user

import (
	"context"
	"errors"
	"strings"

	"github.com/forum-framework/api/internal/shared/apperr"
	"github.com/forum-framework/api/internal/shared/security"
	"github.com/forum-framework/api/internal/shared/validate"
)

// Service implements user/profile business logic.
type Service struct {
	repo *Repository
}

// NewService builds a user Service.
func NewService(repo *Repository) *Service { return &Service{repo: repo} }

// PublicProfile returns the public view of a user by username.
func (s *Service) PublicProfile(ctx context.Context, username string) (*PublicProfile, error) {
	u, err := s.repo.GetByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		return nil, err
	}
	p := u.ToPublic()
	return &p, nil
}

// Account returns the private account view by id.
func (s *Service) Account(ctx context.Context, id string) (*Account, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	a := u.ToAccount()
	return &a, nil
}

// UpdateProfile applies a partial update to the editable profile fields.
// Avatar and banner are untouched (they have their own upload endpoints).
func (s *Service) UpdateProfile(ctx context.Context, id string, in UpdateProfileInput) (*Account, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	displayName, bio, err := mergeProfile(current, in)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateProfile(ctx, id, displayName, bio); err != nil {
		return nil, err
	}
	return s.Account(ctx, id)
}

// mergeProfile resolves the new display name and bio: omitted fields keep the
// current value, an empty bio clears it.
func mergeProfile(current *User, in UpdateProfileInput) (string, *string, error) {
	displayName := current.DisplayName
	if in.DisplayName != nil {
		displayName = strings.TrimSpace(*in.DisplayName)
		if err := validate.DisplayName(displayName); err != nil {
			return "", nil, validationErr(err)
		}
	}
	bio := current.Bio
	if in.Bio != nil {
		trimmed := strings.TrimSpace(*in.Bio)
		if len([]rune(trimmed)) > 500 {
			return "", nil, validationErr(errors.New("a bio deve ter no máximo 500 caracteres"))
		}
		bio = nil
		if trimmed != "" {
			bio = &trimmed
		}
	}
	return displayName, bio, nil
}

// ChangePassword verifies the current password and sets a new one.
func (s *Service) ChangePassword(ctx context.Context, id string, in ChangePasswordInput) error {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	ok, err := security.VerifyPassword(in.CurrentPassword, u.PasswordHash)
	if err != nil || !ok {
		return apperr.ErrUnauthorized
	}
	if err := validate.Password(in.NewPassword); err != nil {
		return validationErr(err)
	}
	hash, err := security.HashPassword(in.NewPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, id, hash)
}

// ChangeEmail verifies the password and sets a new, unique email.
func (s *Service) ChangeEmail(ctx context.Context, id string, in ChangeEmailInput) (*Account, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if err := validate.Email(email); err != nil {
		return nil, validationErr(err)
	}
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ok, err := security.VerifyPassword(in.Password, u.PasswordHash)
	if err != nil || !ok {
		return nil, apperr.ErrUnauthorized
	}
	if email != u.Email {
		taken, err := s.repo.ExistsByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, apperr.ErrConflict
		}
	}
	if err := s.repo.UpdateEmail(ctx, id, email); err != nil {
		return nil, err
	}
	return s.Account(ctx, id)
}

func validationErr(err error) error {
	return &validationError{msg: err.Error()}
}

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }
func (e *validationError) Unwrap() error { return apperr.ErrValidation }

// SetAvatar updates the avatar URL and returns the refreshed account.
func (s *Service) SetAvatar(ctx context.Context, id, url string) (*Account, error) {
	if err := s.repo.UpdateAvatar(ctx, id, &url); err != nil {
		return nil, err
	}
	return s.Account(ctx, id)
}

// SetBanner updates the banner URL and returns the refreshed account.
func (s *Service) SetBanner(ctx context.Context, id, url string) (*Account, error) {
	if err := s.repo.UpdateBanner(ctx, id, &url); err != nil {
		return nil, err
	}
	return s.Account(ctx, id)
}
