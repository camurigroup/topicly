// Package httpx holds small HTTP request helpers.
package httpx

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

// DecodeJSON decodes the request body into dst. It rejects unknown fields and
// oversized bodies. Returns an error the caller maps to a 400 response.
func DecodeJSON(r *http.Request, dst any, maxBytes int64) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxBytes))
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
