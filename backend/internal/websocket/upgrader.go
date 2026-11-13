package websocket

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"streamforge-backend/internal/config"
)

// Upgrader wraps gorilla upgrader with application config
type Upgrader struct {
	gorillaUpgrader websocket.Upgrader
}

// NewUpgrader initializes a gorilla WebSocket upgrader with origin checking
func NewUpgrader(cfg *config.Config) Upgrader {
	return Upgrader{
		gorillaUpgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 16384, // Optimized for video chunk streaming
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					// Allow tools without Origin header
					return true
				}

				for _, allowed := range cfg.AllowedOrigins {
					if allowed == "*" || strings.EqualFold(allowed, origin) {
						return true
					}
				}
				return false
			},
		},
	}
}

// GetGorillaUpgrader exposes the underlying gorilla websocket upgrader
func (u *Upgrader) GetGorillaUpgrader() websocket.Upgrader {
	return u.gorillaUpgrader
}

