// Package middleware provides HTTP middleware for the Brewly backend.
package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/kelvinandreas/brewly/internal/domain"
	jwtpkg "github.com/kelvinandreas/brewly/pkg/jwt"
	"github.com/kelvinandreas/brewly/pkg/response"
)

// RequireAuth returns middleware that validates the Bearer access token and
// enforces that the caller holds one of the allowed roles.
// Pass no roles to allow any authenticated user.
func RequireAuth(secret string, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "missing or malformed Authorization header")
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")

			claims, err := jwtpkg.Verify(token, secret)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
				return
			}

			if len(roles) > 0 && !slices.Contains(roles, claims.Role) {
				response.Error(w, http.StatusForbidden, "forbidden", "insufficient role")
				return
			}

			ctx := context.WithValue(r.Context(), domain.ContextKeyUserID, claims.Sub)
			ctx = context.WithValue(ctx, domain.ContextKeyRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromCtx extracts the authenticated user's UUID string from context.
func UserIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(domain.ContextKeyUserID).(string)
	return v
}

// RoleFromCtx extracts the authenticated user's role from context.
func RoleFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(domain.ContextKeyRole).(string)
	return v
}
