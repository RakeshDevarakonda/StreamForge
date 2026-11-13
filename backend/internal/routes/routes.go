package routes

import (
	"net/http"
	"strings"

	"streamforge-backend/internal/handler"
	"streamforge-backend/internal/service"
	"streamforge-backend/internal/streamer"
	ws "streamforge-backend/internal/websocket"
)

// Register configures all application routes on the provided ServeMux
func Register(
	mux *http.ServeMux,
	store *service.StreamStore,
	manager *streamer.StreamManager,
	upgrader ws.Upgrader,
) {
	streamH := handler.NewStreamHandler(store, manager)
	wsH := handler.NewWSHandler(upgrader.GetGorillaUpgrader(), manager)

	// Health check route
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"service": "streamforge-backend",
		})
	})

	// Stream collection routes
	mux.HandleFunc("GET /api/streams", streamH.ListStreams)
	mux.HandleFunc("POST /api/streams", streamH.CreateStream)

	// Specific stream item routes (handling GET and DELETE)
	mux.HandleFunc("/api/streams/", func(w http.ResponseWriter, r *http.Request) {
		// Verify there is an ID after /api/streams/
		id := strings.TrimPrefix(r.URL.Path, "/api/streams/")
		if id == "" {
			handler.WriteError(w, http.StatusNotFound, "not found")
			return
		}

		switch r.Method {
		case http.MethodGet:
			streamH.GetStream(w, r)
		case http.MethodDelete:
			streamH.DeleteStream(w, r)
		default:
			handler.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	// Live video WebSocket route
	mux.HandleFunc("/ws/stream/", wsH.ServeWS)
}
