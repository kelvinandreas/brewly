package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/your-handle/brewly/internal/usecase"
	"github.com/your-handle/brewly/pkg/response"
)

// ReportHandler handles /api/reports endpoints.
type ReportHandler struct {
	uc *usecase.ReportUsecase
}

// NewReportHandler constructs a ReportHandler.
func NewReportHandler(uc *usecase.ReportUsecase) *ReportHandler {
	return &ReportHandler{uc: uc}
}

// Revenue GET /api/reports/revenue?granularity=day|week|month&from=RFC3339&to=RFC3339
func (h *ReportHandler) Revenue(w http.ResponseWriter, r *http.Request) {
	granularity := r.URL.Query().Get("granularity")
	if granularity == "" {
		granularity = "day"
	}
	from, to, ok := parseDateRange(w, r)
	if !ok {
		return
	}
	rows, err := h.uc.Revenue(r.Context(), granularity, from, to)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not load revenue report")
		return
	}
	response.OK(w, map[string]any{"rows": rows, "granularity": granularity, "from": from, "to": to})
}

// BestSellers GET /api/reports/best-sellers?from=RFC3339&to=RFC3339&limit=10
func (h *ReportHandler) BestSellers(w http.ResponseWriter, r *http.Request) {
	from, to, ok := parseDateRange(w, r)
	if !ok {
		return
	}
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		n, err := strconv.Atoi(l)
		if err != nil || n <= 0 {
			response.Error(w, http.StatusBadRequest, "bad_request", "invalid limit")
			return
		}
		limit = n
	}
	rows, err := h.uc.BestSellers(r.Context(), from, to, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not load best-sellers report")
		return
	}
	response.OK(w, map[string]any{"rows": rows})
}

// HourlyVolume GET /api/reports/hourly-volume?date=2006-01-02
func (h *ReportHandler) HourlyVolume(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "date is required (YYYY-MM-DD)")
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid date format (YYYY-MM-DD required)")
		return
	}
	rows, err := h.uc.HourlyVolume(r.Context(), date)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "could not load hourly volume report")
		return
	}
	response.OK(w, map[string]any{"rows": rows, "date": dateStr})
}

func parseDateRange(w http.ResponseWriter, r *http.Request) (time.Time, time.Time, bool) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "from and to are required (RFC3339)")
		return time.Time{}, time.Time{}, false
	}
	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid from date (RFC3339 required)")
		return time.Time{}, time.Time{}, false
	}
	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid to date (RFC3339 required)")
		return time.Time{}, time.Time{}, false
	}
	return from, to, true
}
