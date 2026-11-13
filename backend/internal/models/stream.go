package models

import (
	"net/url"
	"time"
)

// StreamStatus represents the lifecycle state of an RTSP stream
type StreamStatus string

const (
	StatusConnecting StreamStatus = "connecting"
	StatusLive       StreamStatus = "live"
	StatusError      StreamStatus = "error"
	StatusStopped    StreamStatus = "stopped"
)

// Stream represents an active or registered RTSP stream
type Stream struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	RTSPURL      string       `json:"rtsp_url"`
	Status       StreamStatus `json:"status"`
	ErrorMessage string       `json:"error_message,omitempty"`
	ViewerCount  int          `json:"viewer_count"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// Sanitize returns a shallow copy of the stream with any RTSP credentials masked
func (s *Stream) Sanitize() *Stream {
	cp := *s
	cp.RTSPURL = MaskRTSPCredentials(s.RTSPURL)
	return &cp
}

// MaskRTSPCredentials masks username/passwords in RTSP URLs (e.g. rtsp://user:pass@host -> rtsp://***:***@host)
func MaskRTSPCredentials(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if u.User != nil {
		u.User = url.UserPassword("***", "***")
		return u.String()
	}
	return rawURL
}

// CreateStreamRequest defines the payload for creating/registering a new stream
type CreateStreamRequest struct {
	Name    string `json:"name"`
	RTSPURL string `json:"rtsp_url"`
}

