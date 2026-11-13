package streamer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	"streamforge-backend/internal/models"
)

// safeBuffer is a thread-safe bytes buffer for capturing stderr
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// FFmpegProcess manages the execution and output piping of an FFmpeg transcoding instance
type FFmpegProcess struct {
	cmd        *exec.Cmd
	ffmpegPath string
	rtspURL    string
	cancelFunc context.CancelFunc
	stdoutPipe io.ReadCloser
	stderrBuf  *safeBuffer
	mu         sync.Mutex
	isRunning  bool
}

// NewFFmpegProcess creates an FFmpeg process configured for low-latency RTSP to MPEG1 conversion
func NewFFmpegProcess(ffmpegPath, rtspURL string) *FFmpegProcess {
	return &FFmpegProcess{
		ffmpegPath: ffmpegPath,
		rtspURL:    rtspURL,
		stderrBuf:  &safeBuffer{},
	}
}

// Start launches the FFmpeg process and returns the stdout stream
func (p *FFmpegProcess) Start(ctx context.Context) (io.ReadCloser, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning {
		return p.stdoutPipe, nil
	}

	cmdCtx, cancel := context.WithCancel(ctx)
	p.cancelFunc = cancel

	// Low-latency FFmpeg parameters for RTSP over TCP -> MPEG1 video stream over MPEG-TS
	// Note: Removed -preset and -tune which are specific to libx264/libx265 and invalid for mpeg1video
	args := []string{
		"-rtsp_transport", "tcp",         // Force TCP for reliable packet delivery
		"-timeout", "5000000",              // 5 seconds network timeout
		"-i", p.rtspURL,                    // Input RTSP stream URL
		"-f", "mpegts",                     // Container format: MPEG-TS
		"-codec:v", "mpeg1video",           // Video codec: MPEG1 (supported natively by JSMpeg in canvas)
		"-s", "960x540",                    // Scale resolution for smooth multi-grid rendering
		"-b:v", "1000k",                    // Video bitrate: 1000 kbps
		"-bf", "0",                         // No B-frames to minimize latency
		"-r", "25",                         // Frame rate: 25 fps
		"-an",                              // Disable audio
		"-muxdelay", "0.001",               // Minimum mux delay
		"pipe:1",                           // Output directly to stdout
	}

	p.cmd = exec.CommandContext(cmdCtx, p.ffmpegPath, args...)

	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	p.stdoutPipe = stdout

	// Capture stderr safely
	p.cmd.Stderr = p.stderrBuf

	if err := p.cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	p.isRunning = true
	sanitizedURL := models.MaskRTSPCredentials(p.rtspURL)
	log.Printf("[FFmpeg] Started worker for RTSP: %s (PID: %d)", sanitizedURL, p.cmd.Process.Pid)
	return stdout, nil
}

// Stop gracefully terminates the FFmpeg process and reaps it
func (p *FFmpegProcess) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isRunning {
		return nil
	}

	p.isRunning = false
	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	if p.stdoutPipe != nil {
		_ = p.stdoutPipe.Close()
	}

	if p.cmd != nil && p.cmd.Process != nil {
		done := make(chan error, 1)
		go func() {
			done <- p.cmd.Wait()
		}()

		select {
		case <-done:
		case <-time.After(1500 * time.Millisecond):
			_ = p.cmd.Process.Kill()
			<-done // Ensure process is reaped
		}
		sanitizedURL := models.MaskRTSPCredentials(p.rtspURL)
		log.Printf("[FFmpeg] Stopped worker for: %s", sanitizedURL)
	}

	return nil
}

// LastStderr returns the most recent sanitized error output from FFmpeg
func (p *FFmpegProcess) LastStderr() string {
	raw := p.stderrBuf.String()
	sanitized := models.MaskRTSPCredentials(raw)
	lines := strings.Split(sanitized, "\n")
	if len(lines) > 5 {
		return strings.Join(lines[len(lines)-5:], " ")
	}
	return sanitized
}

// IsRunning reports whether the process is actively executing
func (p *FFmpegProcess) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isRunning
}

