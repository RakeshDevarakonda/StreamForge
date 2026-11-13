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

const (
	IdleTimeout = 30 * time.Second
)

// StreamWorker manages a single RTSP stream's FFmpeg process and fans out chunks to subscribers
type StreamWorker struct {
	streamID    string
	rtspURL     string
	ffmpegPath  string
	process     *FFmpegProcess
	procMu      sync.RWMutex
	subscribers map[*websocket.Client]bool
	subMu       sync.RWMutex
	store       *service.StreamStore
	stopChan    chan struct{}
	lastActive  time.Time
	activeMu    sync.RWMutex
	stopOnce    sync.Once
	isStopped   bool
	onStopped   func(streamID string, w *StreamWorker)
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
			onStopped: func(id string, deadWorker *StreamWorker) {
				m.mu.Lock()
				if current, ok := m.workers[id]; ok && current == deadWorker {
					delete(m.workers, id)
				}
				m.mu.Unlock()
			},
		}
		// Register subscriber FIRST before starting background runner to avoid race
		worker.subscribers[client] = true
		worker.store.SetViewerCount(worker.streamID, 1)

		m.workers[streamID] = worker
		go worker.Run()
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	return worker.AddSubscriber(client)
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
func (w *StreamWorker) AddSubscriber(client *websocket.Client) error {
	w.subMu.Lock()
	w.subscribers[client] = true
	count := len(w.subscribers)
	w.subMu.Unlock()

	w.activeMu.Lock()
	w.lastActive = time.Now()
	w.activeMu.Unlock()

	w.store.SetViewerCount(w.streamID, count)
	log.Printf("[StreamManager] Stream %s: client connected. Viewers: %d", w.streamID, count)
	return nil
}

// RemoveSubscriber unregisters a client and updates viewer counts
func (w *StreamWorker) RemoveSubscriber(client *websocket.Client) {
	w.subMu.Lock()
	delete(w.subscribers, client)
	count := len(w.subscribers)
	w.subMu.Unlock()

	w.activeMu.Lock()
	w.lastActive = time.Now()
	w.activeMu.Unlock()

	w.store.SetViewerCount(w.streamID, count)
	log.Printf("[StreamManager] Stream %s: client disconnected. Viewers: %d", w.streamID, count)
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
				log.Printf("[StreamManager] Stream %s has 0 viewers for %v. Pausing worker.", w.streamID, IdleTimeout)
				w.store.UpdateStatus(w.streamID, models.StatusStopped, "")
				w.Stop()
				return
			}

			// Interruptible sleep
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
			log.Printf("[StreamManager] Stream %s FFmpeg start failed: %v", w.streamID, err)
			w.updateStatusUnlessStopped(models.StatusError, err.Error())

			select {
			case <-w.stopChan:
				return
			case <-time.After(backoff):
			}
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		// Pipe loop: read binary packets and fan-out to subscribers
		buf := make([]byte, 8192)
		firstPacketReceived := false

		for {
			select {
			case <-w.stopChan:
				_ = processRef.Stop()
				return
			default:
			}

			// Check idle state during stdout read loop
			w.subMu.RLock()
			currentViewers := len(w.subscribers)
			w.subMu.RUnlock()

			if currentViewers == 0 {
				w.activeMu.RLock()
				idleDuration := time.Since(w.lastActive)
				w.activeMu.RUnlock()
				if idleDuration > IdleTimeout {
					log.Printf("[StreamManager] Stream %s idle timeout reached during read. Stopping.", w.streamID)
					_ = processRef.Stop()
					w.Stop()
					return
				}
			}

			n, readErr := stdout.Read(buf)
			if n > 0 {
				if !firstPacketReceived {
					firstPacketReceived = true
					w.updateStatusUnlessStopped(models.StatusLive, "")
					backoff = 1 * time.Second
				}

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

		_ = processRef.Stop()
		stderrMsg := processRef.LastStderr()
		if stderrMsg != "" {
			w.updateStatusUnlessStopped(models.StatusError, stderrMsg)
		} else {
			w.updateStatusUnlessStopped(models.StatusError, "stream connection lost")
		}

		// Interruptible backoff
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
			// Client's channel is full; drop packet to preserve low latency
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

		w.store.UpdateStatus(w.streamID, models.StatusStopped, "")

		if w.onStopped != nil {
			w.onStopped(w.streamID, w)
		}
	})
}

// updateStatusUnlessStopped ensures an exiting worker does not overwrite STOPPED with ERROR
func (w *StreamWorker) updateStatusUnlessStopped(status models.StreamStatus, errorMsg string) {
	w.procMu.RLock()
	stopped := w.isStopped
	w.procMu.RUnlock()

	if !stopped {
		w.store.UpdateStatus(w.streamID, status, errorMsg)
	}
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
