package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"streamforge-backend/internal/config"
	"streamforge-backend/internal/models"
)

var (
	ErrStreamNotFound = errors.New("stream not found")
	ErrInvalidRTSPURL = errors.New("invalid RTSP URL: must start with rtsp:// or rtsps://")
	ErrStreamExists   = errors.New("a stream with this RTSP URL already exists")
)

// StreamStore manages in-memory cameras/streams with thread safety
type StreamStore struct {
	mu      sync.RWMutex
	streams map[string]*models.Stream
}

// NewStreamStore creates an in-memory store pre-seeded with sample camera feeds
func NewStreamStore(cfg *config.Config) *StreamStore {
	store := &StreamStore{
		streams: make(map[string]*models.Stream),
	}

	baseURL := "rtsp://localhost:8554"
	if cfg != nil && cfg.DefaultRTSPBaseURL != "" {
		baseURL = cfg.DefaultRTSPBaseURL
	}

	// Seed with default demo cameras for immediate out-of-the-box testing
	defaultCameras := []models.Stream{
		{
			ID:          "camera-1",
			Name:        "Office Entrance",
			RTSPURL:     fmt.Sprintf("%s/live/stream1", baseURL),
			Status:      models.StatusStopped,
			ViewerCount: 0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "camera-2",
			Name:        "Warehouse Floor",
			RTSPURL:     fmt.Sprintf("%s/live/stream2", baseURL),
			Status:      models.StatusStopped,
			ViewerCount: 0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, cam := range defaultCameras {
		c := cam
		store.streams[c.ID] = &c
	}

	return store
}

// GetAll returns a copy of all registered streams
func (s *StreamStore) GetAll() []*models.Stream {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Stream, 0, len(s.streams))
	for _, stream := range s.streams {
		cp := *stream
		result = append(result, &cp)
	}
	return result
}

// Get returns a single stream by ID
func (s *StreamStore) Get(id string) (*models.Stream, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stream, exists := s.streams[id]
	if !exists {
		return nil, ErrStreamNotFound
	}

	cp := *stream
	return &cp, nil
}

// Add validates and adds a new stream to the store
func (s *StreamStore) Add(req models.CreateStreamRequest) (*models.Stream, error) {
	rtspURL := strings.TrimSpace(req.RTSPURL)
	if !strings.HasPrefix(rtspURL, "rtsp://") && !strings.HasPrefix(rtspURL, "rtsps://") {
		return nil, ErrInvalidRTSPURL
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicates
	for _, existing := range s.streams {
		if existing.RTSPURL == rtspURL {
			return nil, ErrStreamExists
		}
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = fmt.Sprintf("Camera-%d", len(s.streams)+1)
	}

	id := "cam-" + uuid.New().String()[:8]
	now := time.Now()

	newStream := &models.Stream{
		ID:          id,
		Name:        name,
		RTSPURL:     rtspURL,
		Status:      models.StatusStopped,
		ViewerCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.streams[id] = newStream
	cp := *newStream
	return &cp, nil
}

// Delete removes a stream by ID
func (s *StreamStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.streams[id]; !exists {
		return ErrStreamNotFound
	}

	delete(s.streams, id)
	return nil
}

// UpdateStatus changes the operational status and error message of a stream
func (s *StreamStore) UpdateStatus(id string, status models.StreamStatus, errorMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stream, exists := s.streams[id]; exists {
		stream.Status = status
		stream.ErrorMessage = errorMsg
		stream.UpdatedAt = time.Now()
	}
}

// SetViewerCount updates the number of active WebSocket viewers for a stream
func (s *StreamStore) SetViewerCount(id string, count int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stream, exists := s.streams[id]; exists {
		stream.ViewerCount = count
		stream.UpdatedAt = time.Now()
	}
}
