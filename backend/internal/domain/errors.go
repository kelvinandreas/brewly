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
)
