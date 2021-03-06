package authorization

import (
	"net/http"

	"context"
)

// User is a parsed Authorization header
type User struct {
	// AuthorizationValue is the auth header value with which a user was created
	AuthorizationValue string
	IsDummy            bool
	Permissions        Permissions
	UserID             []string
}

// Parser creates a User from an authorization header
type Parser interface {
	Parse(authHeader string) (*User, error)
}

// Middleware creates a User from a Parser and puts it in the request context
// which can be later obtained by calling to UserFromContext()
// Errors returned from Parser are printed to the response body
func Middleware(p Parser) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, errMessageNoAuthorizationHeader, http.StatusUnauthorized)
				return
			}

			u, err := p.Parse(authHeader)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := ContextWithUser(r.Context(), u)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

const errMessageNoAuthorizationHeader string = "Authorization header required"

// UserFromContext returns the User stored in a context
func UserFromContext(ctx context.Context) (*User, bool) {
	val, ok := ctx.Value(activeUser).(*User)
	return val, ok
}

// ContextWithUser returns a new `context.Context` that holds a reference to the user `u`
func ContextWithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, activeUser, u)
}

// ContextCopyUser lookups for a user in a parent context a copies it into another context. Useful
// when creating background context with the parent's values
func ContextCopyUser(parent, background context.Context) context.Context {
	if u, ok := UserFromContext(parent); ok {
		background = ContextWithUser(background, u)
	}
	return background
}

type contextKey struct{}

var activeUser = contextKey{}
