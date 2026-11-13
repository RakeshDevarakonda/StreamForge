package routes

import (
	"net/http"

	"streamforge-backend/internal/handler"
	"streamforge-backend/internal/streamer"
	ws "streamforge-backend/internal/websocket"
)

// Register configures the clean stateless backend routes
func Register(
	mux *http.ServeMux,
	manager *streamer.StreamManager,
	upgrader ws.Upgrader,
) {
	wsH := handler.NewWSHandler(upgrader.GetGorillaUpgrader(), manager)

	// Health check route
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"service": "streamforge-backend",
		})
	})

	// On-demand Stateless WebSocket streaming route: /ws/live?url=rtsp://...
	mux.HandleFunc("/ws/live", wsH.ServeWS)
	mux.HandleFunc("/ws/stream/", wsH.ServeWS)
}
