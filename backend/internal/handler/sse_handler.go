package handler

import (
	"fmt"
	"net/http"
	"time"

	jwtpkg "github.com/kelvinandreas/brewly/pkg/jwt"
	"github.com/kelvinandreas/brewly/pkg/response"
	"github.com/kelvinandreas/brewly/pkg/sse"
)

// SSEHandler streams Server-Sent Events to connected clients.
type SSEHandler struct {
	kitchenBroker *sse.Broker
	songBroker    *sse.Broker
	accessSecret  string
}

// NewSSEHandler constructs an SSEHandler.
func NewSSEHandler(kitchenBroker, songBroker *sse.Broker, accessSecret string) *SSEHandler {
	return &SSEHandler{
		kitchenBroker: kitchenBroker,
		songBroker:    songBroker,
		accessSecret:  accessSecret,
	}
}

// KitchenStream GET /api/sse/kitchen?token=<access_jwt>
func (h *SSEHandler) KitchenStream(w http.ResponseWriter, r *http.Request) {
	h.stream(w, r, h.kitchenBroker)
}

// SongQueueStream GET /api/sse/song-queue?token=<access_jwt>
func (h *SSEHandler) SongQueueStream(w http.ResponseWriter, r *http.Request) {
	h.stream(w, r, h.songBroker)
}

// stream is the shared SSE pump: validates token, subscribes to broker, writes frames.
func (h *SSEHandler) stream(w http.ResponseWriter, r *http.Request, broker *sse.Broker) {
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

	events, unsubscribe := broker.Subscribe()
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
