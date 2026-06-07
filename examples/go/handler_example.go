//go:build ignore

// Canonical pattern for a Chi handler.
// Rules: parse request → call usecase → write response. Zero business logic.
package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kelvinandreas/brewly/internal/domain"
	"github.com/kelvinandreas/brewly/internal/usecase"
	"github.com/kelvinandreas/brewly/pkg/response"
)

// ThingHandler handles HTTP requests for Things.
type ThingHandler struct {
	uc       *usecase.ThingUsecase
	validate *validator.Validate
}

// NewThingHandler constructs a ThingHandler. Called from cmd/api/main.go.
func NewThingHandler(uc *usecase.ThingUsecase, v *validator.Validate) *ThingHandler {
	return &ThingHandler{uc: uc, validate: v}
}

// RegisterRoutes mounts routes onto a Chi router.
func (h *ThingHandler) RegisterRoutes(r chi.Router) {
	r.Route("/things", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.getByID)
	})
}

type createThingRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

func (h *ThingHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createThingRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(w, err)
		return
	}

	thing, err := h.uc.CreateThing(req.Name)
	if err != nil {
		if errors.Is(err, domain.ErrThingConflict) {
			response.Error(w, http.StatusConflict, "thing_conflict", "thing already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}

	response.OK(w, http.StatusCreated, thing)
}

func (h *ThingHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_id", "id must be a valid UUID")
		return
	}

	thing, err := h.uc.GetThing(id)
	if err != nil {
		if errors.Is(err, domain.ErrThingNotFound) {
			response.Error(w, http.StatusNotFound, "thing_not_found", "thing not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}

	response.OK(w, http.StatusOK, thing)
}

func (h *ThingHandler) list(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not_implemented", "not implemented")
}
