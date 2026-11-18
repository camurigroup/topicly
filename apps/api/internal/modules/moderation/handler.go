package moderation

import (
	"net/http"

	"github.com/forum-framework/api/internal/shared/authctx"
	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
)

const maxBody = 1 << 16

// Handler exposes moderation endpoints.
type Handler struct {
	svc *Service
}

// NewHandler builds the moderation Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func principal(w http.ResponseWriter, r *http.Request) (string, bool) {
	p, ok := authctx.From(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
		return "", false
	}
	return p.UserID, true
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := httpx.DecodeJSON(r, dst, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return false
	}
	return true
}

// Ban handles POST /admin/users/{username}/ban.
func (h *Handler) Ban(w http.ResponseWriter, r *http.Request) { h.sanction(w, r, TypeBan) }

// Mute handles POST /admin/users/{username}/mute.
func (h *Handler) Mute(w http.ResponseWriter, r *http.Request) { h.sanction(w, r, TypeMute) }

func (h *Handler) sanction(w http.ResponseWriter, r *http.Request, typ string) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	var in SanctionInput
	if !decode(w, r, &in) {
		return
	}
	apply := h.svc.Mute
	if typ == TypeBan {
		apply = h.svc.Ban
	}
	s, err := apply(r.Context(), uid, r.PathValue("username"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Created(w, s)
}

// Revoke handles POST /admin/sanctions/{id}/revoke.
func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	var in RevokeInput
	if !decode(w, r, &in) {
		return
	}
	s, err := h.svc.Revoke(r.Context(), uid, r.PathValue("id"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, s)
}

// Warn handles POST /admin/users/{username}/warnings.
func (h *Handler) Warn(w http.ResponseWriter, r *http.Request) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	var in WarningInput
	if !decode(w, r, &in) {
		return
	}
	wn, err := h.svc.Warn(r.Context(), uid, r.PathValue("username"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Created(w, wn)
}

// UserSummary handles GET /admin/users/{username}/moderation.
func (h *Handler) UserSummary(w http.ResponseWriter, r *http.Request) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	sum, err := h.svc.UserSummary(r.Context(), uid, r.PathValue("username"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, sum)
}

// ListSanctions handles GET /admin/sanctions?type=&active=1&page=&limit=.
func (h *Handler) ListSanctions(w http.ResponseWriter, r *http.Request) {
	p := httpx.Paginate(r, 20, 100)
	q := r.URL.Query()
	items, total, err := h.svc.ListSanctions(r.Context(), SanctionFilter{
		Type: q.Get("type"), ActiveOnly: q.Get("active") == "1" || q.Get("active") == "true",
		Q: q.Get("q"), Ascending: q.Get("order") == "asc",
		Limit: p.Limit, Offset: p.Offset,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Paginated(w, items, response.Meta{Total: total, Page: p.Page, Limit: p.Limit})
}

// CreateReport handles POST /reports.
func (h *Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	var in CreateReportInput
	if !decode(w, r, &in) {
		return
	}
	rp, err := h.svc.CreateReport(r.Context(), uid, in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Created(w, rp)
}

// ListReports handles GET /admin/reports?status=&targetType=&page=&limit=.
func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	p := httpx.Paginate(r, 20, 100)
	q := r.URL.Query()
	items, total, err := h.svc.ListReports(r.Context(), ReportFilter{
		Status: q.Get("status"), TargetType: q.Get("targetType"), Q: q.Get("q"), Order: q.Get("order"),
		Limit: p.Limit, Offset: p.Offset,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Paginated(w, items, response.Meta{Total: total, Page: p.Page, Limit: p.Limit})
}

// ReportCounts handles GET /admin/reports/counts.
func (h *Handler) ReportCounts(w http.ResponseWriter, r *http.Request) {
	counts, err := h.svc.ReportCounts(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, counts)
}

// UpdateReport handles PATCH /admin/reports/{id}.
func (h *Handler) UpdateReport(w http.ResponseWriter, r *http.Request) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	var in UpdateReportInput
	if !decode(w, r, &in) {
		return
	}
	rp, err := h.svc.UpdateReport(r.Context(), uid, r.PathValue("id"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, rp)
}

// RequireUnmuted blocks writes (topics, replies, likes, follows) from muted
// users. It must be chained after RequireAuth.
func (h *Handler) RequireUnmuted(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p, ok := authctx.From(r.Context()); ok {
			s, err := h.svc.Restriction(r.Context(), p.UserID, TypeMute)
			if err != nil {
				response.Error(w, http.StatusInternalServerError, "internal_error", "erro interno")
				return
			}
			if s != nil {
				response.Error(w, http.StatusForbidden, "muted", Describe(s))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
