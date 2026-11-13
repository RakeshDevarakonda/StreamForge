package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"streamforge-backend/internal/models"
	"streamforge-backend/internal/service"
	"streamforge-backend/internal/streamer"
)

// StreamHandler handles HTTP requests for managing RTSP streams
type StreamHandler struct {
	store   *service.StreamStore
	manager *streamer.StreamManager
}

// NewStreamHandler creates a new stream handler instance
func NewStreamHandler(store *service.StreamStore, manager *streamer.StreamManager) *StreamHandler {
	return &StreamHandler{
		store:   store,
		manager: manager,
	}
}

// ListStreams returns all registered streams with credentials sanitized
func (h *StreamHandler) ListStreams(w http.ResponseWriter, r *http.Request) {
	streams := h.store.GetAll()
	sanitized := make([]*models.Stream, 0, len(streams))
	for _, s := range streams {
		sanitized = append(sanitized, s.Sanitize())
	}
	WriteJSON(w, http.StatusOK, sanitized)
}

// CreateStream registers a new RTSP stream
func (h *StreamHandler) CreateStream(w http.ResponseWriter, r *http.Request) {
	var req models.CreateStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	req.RTSPURL = strings.TrimSpace(req.RTSPURL)
	if req.RTSPURL == "" {
		WriteError(w, http.StatusBadRequest, "rtsp_url is required")
		return
	}

	stream, err := h.store.Add(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRTSPURL) {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, service.ErrStreamExists) {
			WriteError(w, http.StatusConflict, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, stream.Sanitize())
}

// DeleteStream removes a registered stream and stops any active worker
func (h *StreamHandler) DeleteStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "stream ID is required")
		return
	}

	// Stop any active FFmpeg transcoding process for this stream
	h.manager.StopStream(id)

	if err := h.store.Delete(id); err != nil {
		if errors.Is(err, service.ErrStreamNotFound) {
			WriteError(w, http.StatusNotFound, "stream not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"deleted": true,
		"id":      id,
	})
}

