package service

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"streamforge-backend/internal/models"
)

var (
	ErrStreamNotFound = errors.New("stream not found")
	ErrInvalidRTSPURL = errors.New("invalid RTSP URL: must be a valid rtsp:// or rtsps:// URL with host and port/path")
	ErrStreamExists   = errors.New("a stream with this RTSP URL already exists")
	ErrMaxStreams     = errors.New("maximum number of concurrent streams reached (limit: 16)")
)

const MaxAllowedStreams = 16

// StreamStore manages in-memory cameras/streams with thread safety
type StreamStore struct {
	mu      sync.RWMutex
	streams map[string]*models.Stream
}

// NewStreamStore creates an empty in-memory store for registered cameras/streams
func NewStreamStore() *StreamStore {
	return &StreamStore{
		streams: make(map[string]*models.Stream),
	}
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

	// Robust URL parsing & scheme validation
	parsedURL, err := url.Parse(rtspURL)
	if err != nil {
		return nil, ErrInvalidRTSPURL
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "rtsp" && scheme != "rtsps" {
		return nil, ErrInvalidRTSPURL
	}

	hostname := parsedURL.Hostname()
	if hostname == "" {
		return nil, ErrInvalidRTSPURL
	}

	// SSRF Protection: Disallow AWS/cloud metadata IP (169.254.169.254)
	if hostname == "169.254.169.254" || strings.HasPrefix(hostname, "169.254.") {
		return nil, errors.New("access to cloud link-local metadata addresses is restricted")
	}

	// Ensure port if specified is valid
	if portStr := parsedURL.Port(); portStr != "" {
		if _, err := net.LookupPort("tcp", portStr); err != nil {
			return nil, fmt.Errorf("invalid port in RTSP URL: %w", err)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.streams) >= MaxAllowedStreams {
		return nil, ErrMaxStreams
	}

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
