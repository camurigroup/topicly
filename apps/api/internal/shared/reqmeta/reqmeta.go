// Package reqmeta carries request metadata (client IP) through the context so
// that services can record it (e.g. in audit logs) without depending on HTTP.
package reqmeta

import "context"

type ctxKey int

const ipKey ctxKey = 0

// WithIP returns a context carrying the client IP.
func WithIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ipKey, ip)
}

// IP returns the client IP stored in the context, or "".
func IP(ctx context.Context) string {
	ip, _ := ctx.Value(ipKey).(string)
	return ip
}
