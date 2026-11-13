package models

import "time"

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

// CreateStreamRequest defines the payload for creating/registering a new stream
type CreateStreamRequest struct {
	Name    string `json:"name"`
	RTSPURL string `json:"rtsp_url"`
}

// StreamStats provides live metrics for an active stream
type StreamStats struct {
	StreamID    string       `json:"stream_id"`
	Status      StreamStatus `json:"status"`
	ViewerCount int          `json:"viewer_count"`
	BytesOut    int64        `json:"bytes_out"`
	UptimeSec   int64        `json:"uptime_sec"`
}
