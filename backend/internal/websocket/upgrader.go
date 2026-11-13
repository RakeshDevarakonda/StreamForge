package websocket

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"streamforge-backend/internal/config"
)

// NewUpgrader initializes a gorilla WebSocket upgrader with origin checking
func NewUpgrader(cfg *config.Config) websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 16384, // Optimized for video chunk streaming
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				// Allow tools (e.g. Postman, VLC, CLI scripts) without Origin header
				return true
			}

			for _, allowed := range cfg.AllowedOrigins {
				if allowed == "*" || strings.EqualFold(allowed, origin) {
					return true
				}
			}
			return false
		},
	}
}
