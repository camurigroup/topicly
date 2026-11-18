package category

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/shared/apperr"
	"github.com/forum-framework/api/internal/shared/id"
	"github.com/forum-framework/api/internal/shared/slug"
)

// Service implements category business logic.
type Service struct {
	repo *Repository
}

// NewService builds a category Service.
func NewService(repo *Repository) *Service { return &Service{repo: repo} }

// List returns categories; activeOnly limits to enabled ones.
func (s *Service) List(ctx context.Context, activeOnly bool) ([]Category, error) {
	return s.repo.List(ctx, activeOnly)
}

// GetBySlug returns a category by slug.
func (s *Service) GetBySlug(ctx context.Context, slugValue string) (*Category, error) {
	return s.repo.GetBySlug(ctx, slugValue)
}

// Create validates and stores a new category.
func (s *Service) Create(ctx context.Context, in CreateInput) (*Category, error) {
	name := strings.TrimSpace(in.Name)
	if l := len([]rune(name)); l < 2 || l > 120 {
		return nil, validationErr("o nome deve ter entre 2 e 120 caracteres")
	}
	if in.ParentID != nil {
		if _, err := s.repo.GetByID(ctx, *in.ParentID); err != nil {
			if errors.Is(err, apperr.ErrNotFound) {
				return nil, validationErr("categoria pai inexistente")
			}
			return nil, err
		}
	}
	slugValue, err := s.uniqueSlug(ctx, name)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	c := &Category{
		ID:          id.New(),
		ParentID:    in.ParentID,
		Name:        name,
		Slug:        slugValue,
		Description: in.Description,
		Icon:        in.Icon,
		CoverURL:    in.CoverURL,
		Position:    in.Position,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Update edits an existing category.
func (s *Service) Update(ctx context.Context, idValue string, in UpdateInput) (*Category, error) {
	c, err := s.repo.GetByID(ctx, idValue)
	if err != nil {
		return nil, err
	}
	if err := applyUpdate(c, in); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// applyUpdate merges a partial update into c: omitted fields are kept, an
// empty optional text clears it. Toggling isActive never wipes the rest.
func applyUpdate(c *Category, in UpdateInput) error {
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if l := len([]rune(name)); l < 2 || l > 120 {
			return validationErr("o nome deve ter entre 2 e 120 caracteres")
		}
		c.Name = name
	}
	c.Description = mergeText(c.Description, in.Description)
	c.Icon = mergeText(c.Icon, in.Icon)
	c.CoverURL = mergeText(c.CoverURL, in.CoverURL)
	if in.Position != nil {
		c.Position = *in.Position
	}
	if in.IsActive != nil {
		c.IsActive = *in.IsActive
	}
	return nil
}

func mergeText(current, next *string) *string {
	if next == nil {
		return current
	}
	v := strings.TrimSpace(*next)
	if v == "" {
		return nil
	}
	return &v
}

// Delete removes a category.
func (s *Service) Delete(ctx context.Context, idValue string) error {
	return s.repo.Delete(ctx, idValue)
}

func (s *Service) uniqueSlug(ctx context.Context, name string) (string, error) {
	base := slug.Make(name)
	if base == "" {
		base = "categoria"
	}
	candidate := base
	for i := 2; ; i++ {
		exists, err := s.repo.SlugExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
		candidate = base + "-" + strconv.Itoa(i)
	}
}

func validationErr(msg string) error { return &validationError{msg: msg} }

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }
func (e *validationError) Unwrap() error { return apperr.ErrValidation }
