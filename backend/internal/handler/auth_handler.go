// Package handler provides HTTP handlers for the Brewly backend.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
	"github.com/your-handle/brewly/internal/middleware"
	"github.com/your-handle/brewly/internal/usecase"
	"github.com/your-handle/brewly/pkg/response"
)

// AuthHandler handles the five /api/auth endpoints.
type AuthHandler struct {
	uc *usecase.AuthUsecase
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(uc *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

// RegisterOwner POST /api/auth/register-owner
func (h *AuthHandler) RegisterOwner(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if body.Email == "" || body.Password == "" || body.Name == "" {
		response.ValidationError(w, []response.FieldError{
			{Field: "email", Message: "required"},
			{Field: "password", Message: "required"},
			{Field: "name", Message: "required"},
		})
		return
	}

	user, err := h.uc.RegisterOwner(r.Context(), body.Email, body.Password, body.Name)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOwnerExists):
			response.Error(w, http.StatusConflict, "owner_exists", "an owner account already exists")
		case errors.Is(err, domain.ErrEmailTaken):
			response.Error(w, http.StatusConflict, "email_taken", "email already in use")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "could not create owner account")
		}
		return
	}
	response.Created(w, map[string]any{"user": user})
}

// Login POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	user, pair, err := h.uc.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "invalid_credentials", "email or password incorrect")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "login failed")
		return
	}

	setRefreshCookie(w, pair.RefreshToken)
	response.OK(w, map[string]any{
		"accessToken": pair.AccessToken,
		"user":        user,
	})
}

// Refresh POST /api/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(domain.RefreshCookieName)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "missing refresh cookie")
		return
	}

	pair, err := h.uc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		clearRefreshCookie(w)
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid or expired refresh token")
		return
	}

	setRefreshCookie(w, pair.RefreshToken)
	response.OK(w, map[string]any{"accessToken": pair.AccessToken})
}

// Logout POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.UserIDFromCtx(r.Context())
	if userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			_ = h.uc.Logout(r.Context(), userID)
		}
	}
	clearRefreshCookie(w)
	response.NoContent(w)
}

// Me GET /api/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.UserIDFromCtx(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid user id in token")
		return
	}

	user, err := h.uc.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "user_not_found", "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not fetch user")
		return
	}
	response.OK(w, map[string]any{"user": user})
}

// setRefreshCookie writes the httpOnly refresh cookie.
func setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     domain.RefreshCookieName,
		Value:    token,
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
}

// clearRefreshCookie expires the refresh cookie immediately.
func clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     domain.RefreshCookieName,
		Value:    "",
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
