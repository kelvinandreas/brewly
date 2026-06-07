package handler

import (
	"fmt"
	"net/http"
	"time"

	jwtpkg "github.com/your-handle/brewly/pkg/jwt"
	"github.com/your-handle/brewly/pkg/response"
	"github.com/your-handle/brewly/pkg/sse"
)

// SSEHandler streams Server-Sent Events to connected clients.
type SSEHandler struct {
	kitchenBroker *sse.Broker
	accessSecret  string
}

// NewSSEHandler constructs an SSEHandler.
func NewSSEHandler(kitchenBroker *sse.Broker, accessSecret string) *SSEHandler {
	return &SSEHandler{kitchenBroker: kitchenBroker, accessSecret: accessSecret}
}

// KitchenStream GET /api/sse/kitchen?token=<access_jwt>
// Streams order events to kitchen/cashier/owner clients.
// EventSource cannot send Authorization headers, so the access token is passed
// as a query parameter instead.
func (h *SSEHandler) KitchenStream(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "token required")
		return
	}
	if _, err := jwtpkg.Verify(token, h.accessSecret); err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "internal_error", "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	events, unsubscribe := h.kitchenBroker.Subscribe()
	defer unsubscribe()

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-keepAlive.C:
			fmt.Fprintf(w, ": keep-alive\n\n")
			flusher.Flush()

		case e, ok := <-events:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.Type, string(e.Payload))
			flusher.Flush()
		}
	}
}
