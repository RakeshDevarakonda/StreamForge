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
	"streamforge-backend/internal/service"
	"streamforge-backend/internal/streamer"
	"streamforge-backend/internal/websocket"
)

func main() {
	// 1. Load configuration strictly from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] Configuration error: %v", err)
	}

	log.Printf("[Init] Starting StreamForge backend (ENV=%s, PORT=%s)", cfg.Env, cfg.Port)

	// 2. Initialize in-memory stream store (starts completely empty)
	store := service.NewStreamStore()
	log.Println("[Init] Initialized clean stream store")

	// 3. Initialize stream manager for orchestrating FFmpeg child processes
	manager := streamer.NewStreamManager(cfg, store)

	// 4. Initialize WebSocket upgrader
	wsUpgrader := websocket.NewUpgrader(cfg)

	// 5. Initialize HTTP router & routes
	mux := http.NewServeMux()
	routes.Register(mux, store, manager, wsUpgrader)

	// 6. Chain middlewares: Recoverer -> Logger -> CORS -> Mux Router
	var rootHandler http.Handler = mux
	rootHandler = middleware.CORS(cfg)(rootHandler)
	rootHandler = middleware.Logger(rootHandler)
	rootHandler = middleware.Recoverer(rootHandler)

	// 7. Configure HTTP server
	// Note: WriteTimeout is NOT set globally because it terminates long-lived WebSocket streaming connections
	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           rootHandler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// 8. Graceful shutdown handler
	shutdownDone := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-quit

		log.Println("[Shutdown] Signal received. Shutting down server and FFmpeg workers...")

		// Stop all active FFmpeg background processes immediately
		manager.StopAll()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("[Shutdown] Error during server shutdown: %v", err)
		}
		close(shutdownDone)
	}()

	// 9. Start listening
	log.Printf("[StreamForge] Server listening on http://localhost:%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}

	<-shutdownDone
	log.Println("[Shutdown] StreamForge server stopped gracefully")
}
