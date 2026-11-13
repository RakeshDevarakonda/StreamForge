package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"streamforge-backend/internal/config"
	"streamforge-backend/internal/middleware"
	"streamforge-backend/internal/routes"
	"streamforge-backend/internal/streamer"
	"streamforge-backend/internal/websocket"
)

func main() {
	// 1. Load configuration strictly from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] Configuration error: %v", err)
	}

	log.Printf("[Init] Starting StreamForge stateless backend (ENV=%s, PORT=%s)", cfg.Env, cfg.Port)

	// 2. Initialize stateless on-demand stream manager
	manager := streamer.NewStreamManager(cfg)

	// 3. Initialize WebSocket upgrader
	wsUpgrader := websocket.NewUpgrader(cfg)

	// 4. Initialize HTTP router & routes
	mux := http.NewServeMux()
	routes.Register(mux, manager, wsUpgrader)

	// 5. Chain middlewares: Recoverer -> Logger -> CORS -> Mux Router
	var rootHandler http.Handler = mux
	rootHandler = middleware.CORS(cfg)(rootHandler)
	rootHandler = middleware.Logger(rootHandler)
	rootHandler = middleware.Recoverer(rootHandler)

	// 6. Configure HTTP server
	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           rootHandler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// 7. Graceful shutdown handler
	shutdownDone := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-quit

		log.Println("[Shutdown] Signal received. Shutting down server and FFmpeg workers...")
		manager.StopAll()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("[Shutdown] Error during server shutdown: %v", err)
		}
		close(shutdownDone)
	}()

	// 8. Start listening
	log.Printf("[StreamForge] Server listening on http://localhost:%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}

	<-shutdownDone
	log.Println("[Shutdown] StreamForge server stopped gracefully")
}
