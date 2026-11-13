package streamer

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"streamforge-backend/internal/config"
	"streamforge-backend/internal/models"
	"streamforge-backend/internal/service"
	"streamforge-backend/internal/websocket"
)

// StreamWorker manages a single RTSP stream's FFmpeg process and fans out chunks to subscribers
type StreamWorker struct {
	streamID     string
	rtspURL      string
	ffmpegPath   string
	process      *FFmpegProcess
	subscribers  map[*websocket.Client]bool
	subMu        sync.RWMutex
	store        *service.StreamStore
	stopChan     chan struct{}
	lastActive   time.Time
	isAlive      bool
	workerMu     sync.Mutex
	stopOnce     sync.Once
	onStopped    func(streamID string)
}

// StreamManager coordinates all active StreamWorkers
type StreamManager struct {
	cfg     *config.Config
	store   *service.StreamStore
	workers map[string]*StreamWorker
	mu      sync.RWMutex
}

// NewStreamManager creates a stream manager instance
func NewStreamManager(cfg *config.Config, store *service.StreamStore) *StreamManager {
	return &StreamManager{
		cfg:     cfg,
		store:   store,
		workers: make(map[string]*StreamWorker),
	}
}

// Subscribe connects a WebSocket client to a stream, automatically starting FFmpeg if not already running
func (m *StreamManager) Subscribe(streamID string, client *websocket.Client) error {
	stream, err := m.store.Get(streamID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	worker, exists := m.workers[streamID]
	if !exists {
		worker = &StreamWorker{
			streamID:    streamID,
			rtspURL:     stream.RTSPURL,
			ffmpegPath:  m.cfg.FFmpegPath,
			subscribers: make(map[*websocket.Client]bool),
			store:       m.store,
			stopChan:    make(chan struct{}),
			lastActive:  time.Now(),
			onStopped: func(id string) {
				m.mu.Lock()
				delete(m.workers, id)
				m.mu.Unlock()
			},
		}
		m.workers[streamID] = worker
		go worker.Run()
	}
	m.mu.Unlock()

	worker.AddSubscriber(client)
	return nil
}

// Unsubscribe removes a client from a stream worker
func (m *StreamManager) Unsubscribe(streamID string, client *websocket.Client) {
	m.mu.RLock()
	worker, exists := m.workers[streamID]
	m.mu.RUnlock()

	if exists {
		worker.RemoveSubscriber(client)
	}
}

// StopStream stops a running worker and terminates its FFmpeg process
func (m *StreamManager) StopStream(streamID string) {
	m.mu.Lock()
	worker, exists := m.workers[streamID]
	if exists {
		delete(m.workers, streamID)
	}
	m.mu.Unlock()

	if exists {
		worker.Stop()
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

// AddSubscriber registers a client and updates viewer counts
func (w *StreamWorker) AddSubscriber(client *websocket.Client) {
	w.subMu.Lock()
	w.subscribers[client] = true
	count := len(w.subscribers)
	w.subMu.Unlock()

	w.store.SetViewerCount(w.streamID, count)
	log.Printf("[StreamManager] Stream %s: client connected. Viewers: %d", w.streamID, count)
}

// RemoveSubscriber unregisters a client and updates viewer counts
func (w *StreamWorker) RemoveSubscriber(client *websocket.Client) {
	w.subMu.Lock()
	delete(w.subscribers, client)
	count := len(w.subscribers)
	w.subMu.Unlock()

	w.store.SetViewerCount(w.streamID, count)
	log.Printf("[StreamManager] Stream %s: client disconnected. Viewers: %d", w.streamID, count)
}

// Run executes the stream pipeline with automatic error handling and reconnection
func (w *StreamWorker) Run() {
	w.workerMu.Lock()
	w.isAlive = true
	w.workerMu.Unlock()

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
			// Idle check: if idle for 30s with 0 viewers, stop FFmpeg to save resources
			if time.Since(w.lastActive) > 30*time.Second {
				log.Printf("[StreamManager] Stream %s has 0 viewers for 30s. Pausing worker.", w.streamID)
				w.store.UpdateStatus(w.streamID, models.StatusStopped, "")
				w.Stop()
				return
			}
			time.Sleep(1 * time.Second)
			continue
		}

		w.lastActive = time.Now()
		w.store.UpdateStatus(w.streamID, models.StatusConnecting, "")

		w.process = NewFFmpegProcess(w.ffmpegPath, w.rtspURL)
		stdout, err := w.process.Start(context.Background())
		if err != nil {
			log.Printf("[StreamManager] Stream %s FFmpeg start failed: %v", w.streamID, err)
			w.store.UpdateStatus(w.streamID, models.StatusError, err.Error())
			time.Sleep(backoff)
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		w.store.UpdateStatus(w.streamID, models.StatusLive, "")
		backoff = 1 * time.Second // Reset backoff on successful connection

		// Pipe loop: read binary packets and fan-out to subscribers
		buf := make([]byte, 8192) // 8KB buffer
		for {
			select {
			case <-w.stopChan:
				_ = w.process.Stop()
				return
			default:
			}

			n, readErr := stdout.Read(buf)
			if n > 0 {
				packet := make([]byte, n)
				copy(packet, buf[:n])
				w.broadcast(packet)
			}

			if readErr != nil {
				if readErr != io.EOF {
					log.Printf("[StreamManager] Stream %s read error: %v", w.streamID, readErr)
				}
				break
			}
		}

		_ = w.process.Stop()
		stderrMsg := w.process.LastStderr()
		if stderrMsg != "" {
			w.store.UpdateStatus(w.streamID, models.StatusError, stderrMsg)
		} else {
			w.store.UpdateStatus(w.streamID, models.StatusError, "stream connection lost")
		}

		// Reconnection backoff timer responds immediately to stop cancellation
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
			// Client's channel is full (slow consumer); drop packet to maintain real-time low latency
		}
	}
}

// Stop cleanly terminates the worker (safe for multiple calls)
func (w *StreamWorker) Stop() {
	w.stopOnce.Do(func() {
		w.workerMu.Lock()
		w.isAlive = false
		w.workerMu.Unlock()

		close(w.stopChan)
		if w.process != nil {
			_ = w.process.Stop()
		}

		w.store.UpdateStatus(w.streamID, models.StatusStopped, "")

		if w.onStopped != nil {
			w.onStopped(w.streamID)
		}
	})
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
