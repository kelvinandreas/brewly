// Package domain defines Brewly's core entities, repository interfaces, errors, and constants.
package domain

import "errors"

// Sentinel errors for the domain layer. Handlers map these to HTTP status codes.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already in use")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrOwnerExists        = errors.New("owner account already exists")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")

	ErrCategoryNotFound    = errors.New("category not found")
	ErrMenuItemNotFound    = errors.New("menu item not found")
	ErrTableNotFound       = errors.New("table not found")
	ErrTableLabelTaken     = errors.New("table label already in use")
	ErrMenuItemUnavailable = errors.New("menu item unavailable")

	ErrOrderNotFound           = errors.New("order not found")
	ErrOrderCancelled          = errors.New("order is already cancelled")
	ErrInvalidStatusTransition = errors.New("invalid order status transition")
	ErrPaymentConflict         = errors.New("payment already recorded for this order")

	ErrSongRequestNotFound         = errors.New("song request not found")
	ErrSongRequestRateLimited      = errors.New("too many active song requests")
	ErrInvalidSongStatusTransition = errors.New("invalid song request status transition")
	ErrSongAlreadyPlaying          = errors.New("another song is already playing")
)
