// Package httpx holds small HTTP request helpers.
package httpx

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"unicode/utf8"
)

// ErrInvalidUTF8 is returned for bodies that are not valid UTF-8. Go's JSON
// decoder would otherwise silently replace the bad bytes with U+FFFD and the
// text would be stored corrupted (e.g. "AnFFFDncios").
var ErrInvalidUTF8 = errors.New("o texto enviado não está em UTF-8")

// DecodeJSON decodes the request body into dst. It rejects unknown fields,
// oversized bodies and invalid UTF-8. Returns an error the caller maps to 400.
func DecodeJSON(r *http.Request, dst any, maxBytes int64) error {
	body, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, maxBytes))
	if err != nil {
		return err
	}
	if !utf8.Valid(body) {
		return ErrInvalidUTF8
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// BearerToken extracts a bearer token from the Authorization header.
func BearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	return ""
}

// ClientIP returns the best-effort client IP, honoring X-Forwarded-For.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// UserAgent returns a length-capped User-Agent string.
func UserAgent(r *http.Request) string {
	ua := r.Header.Get("User-Agent")
	if len(ua) > 255 {
		return ua[:255]
	}
	return ua
}

// Pagination holds parsed list pagination parameters.
type Pagination struct {
	Page   int
	Limit  int
	Offset int
}

// Paginate reads page and limit from the query string, applying bounds.
func Paginate(r *http.Request, defaultLimit, maxLimit int) Pagination {
	page := atoiDefault(r.URL.Query().Get("page"), 1)
	if page < 1 {
		page = 1
	}
	limit := atoiDefault(r.URL.Query().Get("limit"), defaultLimit)
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return Pagination{Page: page, Limit: limit, Offset: (page - 1) * limit}
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}
