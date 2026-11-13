package handler

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"streamforge-backend/internal/streamer"
	ws "streamforge-backend/internal/websocket"

	"github.com/gorilla/websocket"
)

// WSHandler handles on-demand WebSocket connections for live RTSP streaming
type WSHandler struct {
	upgrader websocket.Upgrader
	manager  *streamer.StreamManager
}

// NewWSHandler creates a new WebSocket handler instance
func NewWSHandler(upgrader websocket.Upgrader, manager *streamer.StreamManager) *WSHandler {
	return &WSHandler{
		upgrader: upgrader,
		manager:  manager,
	}
}

// ServeWS upgrades the HTTP connection and streams video on-demand from the provided RTSP URL
func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// 1. Extract RTSP URL from query parameter: /ws/live?url=rtsp://...
	rawURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if rawURL == "" {
		// Fallback for paths: /ws/stream/{id}
		rawURL = strings.TrimPrefix(r.URL.Path, "/ws/stream/")
		rawURL = strings.Trim(rawURL, "/")
	}

	if rawURL == "" {
		WriteError(w, http.StatusBadRequest, "rtsp url query parameter (?url=rtsp://...) is required")
		return
	}

	// Validate RTSP URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil || (parsedURL.Scheme != "rtsp" && parsedURL.Scheme != "rtsps") || parsedURL.Hostname() == "" {
		WriteError(w, http.StatusBadRequest, "invalid RTSP URL: must start with rtsp:// or rtsps://")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade failed: %v", err)
		return
	}

	client := ws.NewClient(conn)

	// Subscribe client to stream manager using the RTSP URL directly
	if err := h.manager.SubscribeURL(rawURL, client); err != nil {
		log.Printf("[WebSocket] Failed to start stream for %s: %v", rawURL, err)
		_ = conn.Close()
		return
	}

	cleanup := func() {
		h.manager.UnsubscribeURL(rawURL, client)
	}

	go client.WritePump(cleanup)
	client.ReadPump(cleanup)
}
