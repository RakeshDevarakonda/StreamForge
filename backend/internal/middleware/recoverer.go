package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

// Recoverer catches panics during request handling and returns a clean 500 JSON response
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				log.Printf("[PANIC RECOVERED] %v\nStack trace:\n%s", rvr, debug.Stack())

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Internal server error occurred",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
