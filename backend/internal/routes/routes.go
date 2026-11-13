package routes

import (
	"net/http"

	"streamforge-backend/internal/handler"
	"streamforge-backend/internal/service"
	"streamforge-backend/internal/streamer"
	ws "streamforge-backend/internal/websocket"
)

// Register configures all application routes cleanly using Go 1.22+ ServeMux routing
func Register(
	mux *http.ServeMux,
	store *service.StreamStore,
	manager *streamer.StreamManager,
	upgrader ws.Upgrader,
) {
	streamH := handler.NewStreamHandler(store, manager)
	wsH := handler.NewWSHandler(store, upgrader.GetGorillaUpgrader(), manager)

	// Health check route
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"service": "streamforge-backend",
		})
	})

	// Stream REST endpoints
	mux.HandleFunc("GET /api/streams", streamH.ListStreams)
	mux.HandleFunc("POST /api/streams", streamH.CreateStream)
	mux.HandleFunc("DELETE /api/streams/{id}", streamH.DeleteStream)

	// Live video WebSocket endpoint
	mux.HandleFunc("/ws/stream/{id}", wsH.ServeWS)
}
