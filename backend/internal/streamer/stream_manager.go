package streamer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"sync"
	"time"

	"streamforge-backend/internal/config"
	"streamforge-backend/internal/models"
	"streamforge-backend/internal/websocket"
)

const (
	IdleTimeout = 30 * time.Second
)

// StreamWorker manages a single RTSP stream's FFmpeg process and fans out chunks to subscribers
type StreamWorker struct {
	streamKey   string
	rtspURL     string
	ffmpegPath  string
	process     *FFmpegProcess
	procMu      sync.RWMutex
	subscribers map[*websocket.Client]bool
	subMu       sync.RWMutex
	stopChan    chan struct{}
	lastActive  time.Time
	activeMu    sync.RWMutex
	stopOnce    sync.Once
	isStopped   bool
	onStopped   func(streamKey string, w *StreamWorker)
}

// StreamManager coordinates on-demand StreamWorkers
type StreamManager struct {
	cfg     *config.Config
	workers map[string]*StreamWorker
	mu      sync.RWMutex
}

// NewStreamManager creates a stream manager instance
func NewStreamManager(cfg *config.Config) *StreamManager {
	return &StreamManager{
		cfg:     cfg,
		workers: make(map[string]*StreamWorker),
	}
}

// SubscribeURL connects a WebSocket client directly to an RTSP URL on demand
func (m *StreamManager) SubscribeURL(rtspURL string, client *websocket.Client) error {
	key := hashURL(rtspURL)

	m.mu.Lock()
	worker, exists := m.workers[key]
	if !exists {
		worker = &StreamWorker{
			streamKey:   key,
			rtspURL:     rtspURL,
			ffmpegPath:  m.cfg.FFmpegPath,
			subscribers: make(map[*websocket.Client]bool),
			stopChan:    make(chan struct{}),
			lastActive:  time.Now(),
			onStopped: func(k string, deadWorker *StreamWorker) {
				m.mu.Lock()
				if current, ok := m.workers[k]; ok && current == deadWorker {
					delete(m.workers, k)
				}
				m.mu.Unlock()
			},
		}
		worker.subscribers[client] = true
		m.workers[key] = worker
		go worker.Run()
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	return worker.AddSubscriber(client)
}

// UnsubscribeURL removes a client from a stream worker by RTSP URL
func (m *StreamManager) UnsubscribeURL(rtspURL string, client *websocket.Client) {
	key := hashURL(rtspURL)
	m.mu.RLock()
	worker, exists := m.workers[key]
	m.mu.RUnlock()

	if exists {
		worker.RemoveSubscriber(client)
	}
}

// StopAll stops all active stream workers during server shutdown
func (m *StreamManager) StopAll() {
	m.mu.Lock()
	workers := make([]*StreamWorker, 0, len(m.workers))
	for id, w := range m.workers {
		workers = append(workers, w)
		delete(m.workers, id)
	}
	m.mu.Unlock()

	for _, w := range workers {
		w.Stop()
	}
}

// AddSubscriber registers a client
func (w *StreamWorker) AddSubscriber(client *websocket.Client) error {
	w.subMu.Lock()
	w.subscribers[client] = true
	count := len(w.subscribers)
	w.subMu.Unlock()

	w.activeMu.Lock()
	w.lastActive = time.Now()
	w.activeMu.Unlock()

	log.Printf("[StreamManager] Client connected to %s. Viewers: %d", models.MaskRTSPCredentials(w.rtspURL), count)
	return nil
}

// RemoveSubscriber unregisters a client
func (w *StreamWorker) RemoveSubscriber(client *websocket.Client) {
	w.subMu.Lock()
	delete(w.subscribers, client)
	count := len(w.subscribers)
	w.subMu.Unlock()

	w.activeMu.Lock()
	w.lastActive = time.Now()
	w.activeMu.Unlock()

	log.Printf("[StreamManager] Client disconnected from %s. Viewers: %d", models.MaskRTSPCredentials(w.rtspURL), count)
}

// Run executes the stream pipeline with automatic error handling, true LIVE transition on data, and idle timeout
func (w *StreamWorker) Run() {
	backoff := 1 * time.Second
	maxBackoff := 8 * time.Second

	for {
		select {
		case <-w.stopChan:
			return
		default:
		}

		// Check if anyone is still viewing
		w.subMu.RLock()
		activeCount := len(w.subscribers)
		w.subMu.RUnlock()

		if activeCount == 0 {
			w.activeMu.RLock()
			idleDuration := time.Since(w.lastActive)
			w.activeMu.RUnlock()

			if idleDuration > IdleTimeout {
				log.Printf("[StreamManager] Stream %s has 0 viewers for %v. Pausing worker.", models.MaskRTSPCredentials(w.rtspURL), IdleTimeout)
				w.Stop()
				return
			}

			select {
			case <-w.stopChan:
				return
			case <-time.After(1 * time.Second):
			}
			continue
		}

		w.activeMu.Lock()
		w.lastActive = time.Now()
		w.activeMu.Unlock()

		w.procMu.Lock()
		if w.isStopped {
			w.procMu.Unlock()
			return
		}
		w.process = NewFFmpegProcess(w.ffmpegPath, w.rtspURL)
		processRef := w.process
		w.procMu.Unlock()

		stdout, err := processRef.Start(context.Background())
		if err != nil {
			log.Printf("[StreamManager] Stream %s FFmpeg start failed: %v", models.MaskRTSPCredentials(w.rtspURL), err)

			select {
			case <-w.stopChan:
				return
			case <-time.After(backoff):
			}
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		buf := make([]byte, 4096)
		for {
			select {
			case <-w.stopChan:
				_ = processRef.Stop()
				return
			default:
			}

			w.subMu.RLock()
			currentViewers := len(w.subscribers)
			w.subMu.RUnlock()

			if currentViewers == 0 {
				w.activeMu.RLock()
				idleDuration := time.Since(w.lastActive)
				w.activeMu.RUnlock()
				if idleDuration > IdleTimeout {
					log.Printf("[StreamManager] Stream %s idle timeout reached during read. Stopping.", models.MaskRTSPCredentials(w.rtspURL))
					_ = processRef.Stop()
					w.Stop()
					return
				}
			}

			n, readErr := stdout.Read(buf)
			if n > 0 {
				packet := make([]byte, n)
				copy(packet, buf[:n])
				w.broadcast(packet)
			}

			if readErr != nil {
				if readErr != io.EOF {
					log.Printf("[StreamManager] Stream %s read error: %v", models.MaskRTSPCredentials(w.rtspURL), readErr)
				}
				break
			}
		}

		_ = processRef.Stop()

		select {
		case <-w.stopChan:
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

// broadcast sends video chunks to all subscribers non-blockingly
func (w *StreamWorker) broadcast(data []byte) {
	w.subMu.RLock()
	defer w.subMu.RUnlock()

	for client := range w.subscribers {
		select {
		case client.Send <- data:
		default:
		}
	}
}

// Stop cleanly terminates the worker (safe for multiple calls)
func (w *StreamWorker) Stop() {
	w.stopOnce.Do(func() {
		w.procMu.Lock()
		w.isStopped = true
		processRef := w.process
		w.procMu.Unlock()

		close(w.stopChan)
		if processRef != nil {
			_ = processRef.Stop()
		}

		if w.onStopped != nil {
			w.onStopped(w.streamKey, w)
		}
	})
}

func hashURL(u string) string {
	h := sha256.Sum256([]byte(u))
	return hex.EncodeToString(h[:8])
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
