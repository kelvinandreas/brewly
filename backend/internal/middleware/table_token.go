package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"github.com/your-handle/brewly/pkg/response"
	"github.com/your-handle/brewly/pkg/tabletoken"
)

// RequireTableToken returns middleware that validates a customer table token and
// enforces that the token version still matches the table's current version in DB.
func RequireTableToken(secret string, tableRepo domain.TableRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "missing table token")
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")

			claims, err := tabletoken.Verify(token, secret)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid or expired table token")
				return
			}

			tableID, err := uuid.Parse(claims.TableID)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "malformed table id in token")
				return
			}

			table, err := tableRepo.FindByID(r.Context(), tableID)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "table not found")
				return
			}
			if table.DeletedAt != nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "table has been removed")
				return
			}
			if table.TokenVersion != claims.TokenVersion {
				response.Error(w, http.StatusUnauthorized, "token_version_mismatch", "QR code has been regenerated — please scan the new code")
				return
			}

			ctx := context.WithValue(r.Context(), domain.ContextKeyTableID, tableID.String())
			ctx = context.WithValue(ctx, domain.ContextKeyTokenJTI, claims.JTI)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TableIDFromCtx extracts the authenticated table's UUID string from context.
func TableIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(domain.ContextKeyTableID).(string)
	return v
}

// TokenJTIFromCtx extracts the table token's JTI from context.
func TokenJTIFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(domain.ContextKeyTokenJTI).(string)
	return v
}
