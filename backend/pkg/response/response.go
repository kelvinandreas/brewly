// Package response provides standard JSON response helpers for HTTP handlers.
package response

import (
	"encoding/json"
	"net/http"
)

type envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Details any    `json:"details,omitempty"`
}

// FieldError describes a single validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func write(w http.ResponseWriter, status int, body envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// OK writes 200 with a data payload.
func OK(w http.ResponseWriter, data any) {
	write(w, http.StatusOK, envelope{Success: true, Data: data})
}

// Created writes 201 with a data payload.
func Created(w http.ResponseWriter, data any) {
	write(w, http.StatusCreated, envelope{Success: true, Data: data})
}

// OKMsg writes 200 with a data payload and an optional human-readable message.
func OKMsg(w http.ResponseWriter, data any, msg string) {
	write(w, http.StatusOK, envelope{Success: true, Data: data, Message: msg})
}

// NoContent writes 204 with an empty success envelope.
func NoContent(w http.ResponseWriter) {
	write(w, http.StatusNoContent, envelope{Success: true})
}

// Error writes a failure response with a machine-readable error code.
func Error(w http.ResponseWriter, status int, code string, msg string) {
	write(w, status, envelope{Success: false, Error: code, Message: msg})
}

// ValidationError writes 422 with field-level validation details.
func ValidationError(w http.ResponseWriter, details []FieldError) {
	write(w, http.StatusUnprocessableEntity, envelope{
		Success: false,
		Error:   "validation_failed",
		Details: details,
	})
}
