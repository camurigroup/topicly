// Package authctx carries the authenticated principal through request context.
package authctx

import "context"

type ctxKey int

const principalKey ctxKey = 0

// Principal is the authenticated caller.
type Principal struct {
	UserID   string
	Username string
}

// With returns a context carrying the principal.
func With(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

// From extracts the principal from the context.
func From(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}
