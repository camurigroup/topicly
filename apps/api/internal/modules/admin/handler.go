package admin

import (
	"net/http"

	"github.com/forum-framework/api/internal/shared/authctx"
	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/id"
	"github.com/forum-framework/api/internal/shared/response"
	"github.com/forum-framework/api/internal/storage"
)

// Handler exposes the admin dashboard endpoints.
type Handler struct {
	svc     *Service
	storage storage.Provider
}

// NewHandler builds the admin Handler.
func NewHandler(svc *Service, store storage.Provider) *Handler {
	return &Handler{svc: svc, storage: store}
}

// Metrics handles GET /admin/metrics.
func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	t, err := h.svc.Totals(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, t)
}

// Series handles GET /admin/metrics/series?metric=users&period=daily.
func (h *Handler) Series(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	period := q.Get("period")
	if period == "" {
		period = PeriodDaily
	}
	pts, err := h.svc.Series(r.Context(), q.Get("metric"), period)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, pts)
}

func paged[T any](w http.ResponseWriter, p httpx.Pagination, items []T, total int64, err error) {
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Paginated(w, items, response.Meta{Total: total, Page: p.Page, Limit: p.Limit})
}

// Users handles GET /admin/users?q=&role=&status=&sort=&order=&page=&limit=.
func (h *Handler) Users(w http.ResponseWriter, r *http.Request) {
	p := httpx.Paginate(r, 20, 100)
	q := r.URL.Query()
	items, total, err := h.svc.ListUsers(r.Context(), UserFilter{
		Q: q.Get("q"), Role: q.Get("role"), Status: q.Get("status"),
		Sort: q.Get("sort"), Order: q.Get("order"), Limit: p.Limit, Offset: p.Offset,
	})
	paged(w, p, items, total, err)
}

// Topics handles GET /admin/topics?q=&category=&flag=&sort=&order=&page=.
func (h *Handler) Topics(w http.ResponseWriter, r *http.Request) {
	p := httpx.Paginate(r, 20, 100)
	q := r.URL.Query()
	items, total, err := h.svc.ListTopics(r.Context(), TopicFilter{
		Q: q.Get("q"), Category: q.Get("category"), Flag: q.Get("flag"),
		Sort: q.Get("sort"), Order: q.Get("order"), Limit: p.Limit, Offset: p.Offset,
	})
	paged(w, p, items, total, err)
}

// Replies handles GET /admin/replies?q=&state=&sort=&order=&page=.
func (h *Handler) Replies(w http.ResponseWriter, r *http.Request) {
	p := httpx.Paginate(r, 20, 100)
	q := r.URL.Query()
	items, total, err := h.svc.ListReplies(r.Context(), ReplyFilter{
		Q: q.Get("q"), State: q.Get("state"), Sort: q.Get("sort"), Order: q.Get("order"),
		Limit: p.Limit, Offset: p.Offset,
	})
	paged(w, p, items, total, err)
}

// Warnings handles GET /admin/warnings?q=&order=&page=.
func (h *Handler) Warnings(w http.ResponseWriter, r *http.Request) {
	p := httpx.Paginate(r, 20, 100)
	q := r.URL.Query()
	items, total, err := h.svc.ListWarnings(r.Context(), WarningFilter{
		Q: q.Get("q"), Order: q.Get("order"), Limit: p.Limit, Offset: p.Offset,
	})
	paged(w, p, items, total, err)
}

// Storage handles GET /admin/storage.
func (h *Handler) Storage(w http.ResponseWriter, r *http.Request) {
	info, err := h.svc.Storage(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, info)
}

// UploadAsset handles POST /admin/site/assets (multipart "file"): stores a
// brand image (logo, favicon, banner) and returns its public URL. The caller
// then saves the URL in the matching site setting.
func (h *Handler) UploadAsset(w http.ResponseWriter, r *http.Request) {
	p, ok := authctx.From(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
		return
	}
	if err := r.ParseMultipartForm(storage.MaxImageBytes); err != nil {
		httpx.WriteValidation(w, "arquivo inválido ou muito grande")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteValidation(w, "campo 'file' é obrigatório")
		return
	}
	defer file.Close()
	if header.Size > storage.MaxImageBytes {
		httpx.WriteValidation(w, "o arquivo excede o limite de 5 MB")
		return
	}
	contentType, ext, err := storage.DetectImage(file)
	if err != nil {
		httpx.WriteValidation(w, err.Error())
		return
	}
	url, err := h.storage.Save(r.Context(), "brand/"+p.UserID[:8]+"-"+id.New()+ext, file, header.Size, contentType)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Created(w, map[string]string{"url": url})
}
