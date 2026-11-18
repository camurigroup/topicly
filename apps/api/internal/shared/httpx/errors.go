package httpx

import (
	"errors"
	"net/http"

	"github.com/forum-framework/api/internal/shared/apperr"
	"github.com/forum-framework/api/internal/shared/response"
)

type mapping struct {
	status  int
	code    string
	message string
}

var sentinelMap = []struct {
	err error
	m   mapping
}{
	{apperr.ErrValidation, mapping{http.StatusUnprocessableEntity, "validation", "dados inválidos"}},
	{apperr.ErrConflict, mapping{http.StatusConflict, "conflict", "conflito de dados"}},
	{apperr.ErrNotFound, mapping{http.StatusNotFound, "not_found", "não encontrado"}},
	{apperr.ErrForbidden, mapping{http.StatusForbidden, "forbidden", "acesso negado"}},
	{apperr.ErrUnauthorized, mapping{http.StatusUnauthorized, "unauthorized", "credenciais inválidas"}},
}

// WriteError maps a domain error to the response envelope. When the error
// carries its own message (e.g. a wrapped validation error), it is preferred.
func WriteError(w http.ResponseWriter, err error) {
	for _, s := range sentinelMap {
		if errors.Is(err, s.err) {
			msg := s.m.message
			if err.Error() != s.err.Error() {
				msg = err.Error()
			}
			response.Error(w, s.m.status, s.m.code, msg)
			return
		}
	}
	response.Error(w, http.StatusInternalServerError, "internal_error", "erro interno")
}

// WriteValidation writes a 400 for malformed request bodies.
func WriteValidation(w http.ResponseWriter, msg string) {
	response.Error(w, http.StatusBadRequest, "bad_request", msg)
}
