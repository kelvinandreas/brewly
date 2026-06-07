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

// Context keys injected by table-token middleware.
const (
	ContextKeyTableID  contextKey = "tableID"
	ContextKeyTokenJTI contextKey = "tokenJTI"
)

// Order status values.
const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusPreparing = "preparing"
	StatusReady     = "ready"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
)

// Order source values.
const (
	SourceCustomerQR = "customer_qr"
	SourceCashierPOS = "cashier_pos"
)

// Payment method values.
const (
	PaymentMethodCash = "cash"
	PaymentMethodQRIS = "qris"
	PaymentMethodCard = "card"
)
