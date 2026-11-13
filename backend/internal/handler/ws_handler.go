package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"streamforge-backend/internal/streamer"
	ws "streamforge-backend/internal/websocket"
)

// WSHandler handles WebSocket connections for live RTSP streaming
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

// ServeWS upgrades the HTTP connection and streams video frames to the client
func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	streamID := r.PathValue("id")
	if streamID == "" {
		// Fallback for paths without wildcards
		streamID = strings.TrimPrefix(r.URL.Path, "/ws/stream/")
		streamID = strings.Trim(streamID, "/")
	}

	if streamID == "" {
		WriteError(w, http.StatusBadRequest, "stream ID is required")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade failed: %v", err)
		return
	}

	client := ws.NewClient(conn)

	// Subscribe client to stream manager (starts FFmpeg if not already running)
	if err := h.manager.Subscribe(streamID, client); err != nil {
		log.Printf("[WebSocket] Failed to subscribe client to stream %s: %v", streamID, err)
		_ = conn.Close()
		return
	}

	// Clean up subscription when client disconnects
	cleanup := func() {
		h.manager.Unsubscribe(streamID, client)
	}

	// Start read and write pumps in goroutines
	go client.WritePump(cleanup)
	client.ReadPump(cleanup)
}
