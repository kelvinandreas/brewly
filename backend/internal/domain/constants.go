package domain

// Staff roles.
const (
	RoleOwner   = "owner"
	RoleCashier = "cashier"
	RoleKitchen = "kitchen"
)

// RefreshCookieName is the httpOnly cookie that carries the refresh token.
const RefreshCookieName = "BrewlyRefresh"

// contextKey is an unexported type for context keys to prevent collisions.
type contextKey string

// Context keys injected by JWT middleware.
const (
	ContextKeyUserID contextKey = "userID"
	ContextKeyRole   contextKey = "role"
)
