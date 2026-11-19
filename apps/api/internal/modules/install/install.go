// Package install implements the first-run installer: it creates the Owner
// account and the forum identity, then locks itself so it never runs again
// without an explicit, authorized reinstall.
package install

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/modules/auth"
	"github.com/forum-framework/api/internal/modules/category"
	"github.com/forum-framework/api/internal/modules/site"
	"github.com/forum-framework/api/internal/shared/apperr"
	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
	"github.com/forum-framework/api/internal/shared/validate"
)

// Version recorded when the installation completes.
const Version = "1.0.0"

// Status describes the installation state.
type Status struct {
	Installed     bool       `json:"installed"`
	Version       string     `json:"version"`
	InstalledAt   *time.Time `json:"installedAt"`
	TokenRequired bool       `json:"tokenRequired"`
}

// CompleteInput is the installer form.
type CompleteInput struct {
	Token       string `json:"token"`
	ForumName   string `json:"forumName"`
	Description string `json:"description"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type inputError string

func (e inputError) Error() string { return string(e) }
func (e inputError) Unwrap() error { return apperr.ErrValidation }

type stateError struct {
	msg  string
	kind error
}

func (e *stateError) Error() string { return e.msg }
func (e *stateError) Unwrap() error { return e.kind }

func validateInput(in CompleteInput) (CompleteInput, error) {
	in.ForumName = strings.TrimSpace(in.ForumName)
	in.Description = strings.TrimSpace(in.Description)
	if l := len([]rune(in.ForumName)); l < 2 || l > 60 {
		return in, inputError("o nome do fórum deve ter entre 2 e 60 caracteres")
	}
	if len([]rune(in.Description)) > 300 {
		return in, inputError("a descrição deve ter no máximo 300 caracteres")
	}
	if err := validate.Username(strings.TrimSpace(in.Username)); err != nil {
		return in, inputError(err.Error())
	}
	if err := validate.Password(in.Password); err != nil {
		return in, inputError(err.Error())
	}
	return in, nil
}

// tokenOK compares the optional INSTALL_TOKEN in constant time.
func tokenOK(expected, given string) bool {
	if expected == "" {
		return true
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(given)) == 1
}

// Service runs the installer.
type Service struct {
	store      *database.Store
	auth       *auth.Service
	site       *site.Service
	categories *category.Service
	token      string
	mu         sync.Mutex
}

// NewService builds the installer. token is the optional INSTALL_TOKEN.
func NewService(store *database.Store, authSvc *auth.Service, siteSvc *site.Service, catSvc *category.Service, token string) *Service {
	return &Service{store: store, auth: authSvc, site: siteSvc, categories: catSvc, token: token}
}

// Status returns the installation state.
func (s *Service) Status(ctx context.Context) (*Status, error) {
	st := &Status{TokenRequired: s.token != ""}
	var at sql.NullTime
	err := s.store.DB.QueryRowContext(ctx, "SELECT installed, version, installed_at FROM installations WHERE id = 1").
		Scan(&st.Installed, &st.Version, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return st, nil
	}
	if err != nil {
		return nil, err
	}
	if at.Valid {
		t := at.Time
		st.InstalledAt = &t
	}
	return st, nil
}

func (s *Service) markInstalled(ctx context.Context) error {
	q := s.store.Rebind("UPDATE installations SET installed = ?, version = ?, installed_at = ? WHERE id = 1")
	_, err := s.store.DB.ExecContext(ctx, q, true, Version, time.Now().UTC())
	return err
}

func (s *Service) userCount(ctx context.Context) (int64, error) {
	var n int64
	err := s.store.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&n)
	return n, err
}

// Reconcile marks databases that already have accounts (created before the
// installer existed) as installed, so they are never offered the wizard.
func (s *Service) Reconcile(ctx context.Context) (bool, error) {
	st, err := s.Status(ctx)
	if err != nil || st.Installed {
		return false, err
	}
	n, err := s.userCount(ctx)
	if err != nil || n == 0 {
		return false, err
	}
	return true, s.markInstalled(ctx)
}

// Complete creates the Owner, applies the forum identity, seeds a first
// category and locks the installer. It returns the Owner's session.
func (s *Service) Complete(ctx context.Context, in CompleteInput, userAgent, ip string) (*auth.AuthResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, err := s.Status(ctx)
	if err != nil {
		return nil, err
	}
	if st.Installed {
		return nil, &stateError{"o fórum já está instalado", apperr.ErrConflict}
	}
	if !tokenOK(s.token, in.Token) {
		return nil, &stateError{"token de instalação inválido", apperr.ErrForbidden}
	}
	in, err = validateInput(in)
	if err != nil {
		return nil, err
	}
	if n, err := s.userCount(ctx); err != nil {
		return nil, err
	} else if n > 0 {
		_ = s.markInstalled(ctx)
		return nil, &stateError{"o fórum já possui contas e foi marcado como instalado", apperr.ErrConflict}
	}

	// The first account becomes the Owner (see auth.Service.Register).
	res, err := s.auth.Register(ctx, auth.RegisterInput{
		Username: in.Username, Email: in.Email, Password: in.Password, DisplayName: in.DisplayName,
	}, userAgent, ip)
	if err != nil {
		return nil, err
	}
	settings := map[string]string{"brand.name": in.ForumName}
	if in.Description != "" {
		settings["brand.description"] = in.Description
	}
	if _, err := s.site.UpdateSettings(ctx, settings); err != nil {
		return nil, err
	}
	desc := "Apresente-se e converse sobre qualquer assunto."
	if _, err := s.categories.Create(ctx, category.CreateInput{Name: "Geral", Description: &desc}); err != nil {
		return nil, err
	}
	if err := s.markInstalled(ctx); err != nil {
		return nil, err
	}
	return res, nil
}

// Handler exposes the installer endpoints.
type Handler struct{ svc *Service }

// NewHandler builds the installer Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Status handles GET /install/status.
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.Status(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, st)
}

// Complete handles POST /install.
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	var in CompleteInput
	if err := httpx.DecodeJSON(r, &in, 1<<16); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	res, err := h.svc.Complete(r.Context(), in, httpx.UserAgent(r), httpx.ClientIP(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Created(w, res)
}

// RequireInstalled blocks a route (public registration) until the installer
// has created the Owner, so nobody can grab the Owner role on a fresh install.
func (h *Handler) RequireInstalled(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st, err := h.svc.Status(r.Context())
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "internal_error", "erro interno")
			return
		}
		if !st.Installed {
			response.Error(w, http.StatusForbidden, "not_installed", "o fórum ainda não foi instalado")
			return
		}
		next.ServeHTTP(w, r)
	})
}
