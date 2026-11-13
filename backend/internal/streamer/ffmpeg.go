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
)

// FFmpegProcess manages the execution and output piping of an FFmpeg transcoding instance
type FFmpegProcess struct {
	cmd        *exec.Cmd
	ffmpegPath string
	rtspURL    string
	cancelFunc context.CancelFunc
	stdoutPipe io.ReadCloser
	stderrBuf  *bytes.Buffer
	mu         sync.Mutex
	isRunning  bool
}

// NewFFmpegProcess creates an FFmpeg process configured for low-latency RTSP to MPEG1 conversion
func NewFFmpegProcess(ffmpegPath, rtspURL string) *FFmpegProcess {
	return &FFmpegProcess{
		ffmpegPath: ffmpegPath,
		rtspURL:    rtspURL,
		stderrBuf:  &bytes.Buffer{},
	}
}

// Start launches the FFmpeg process with low-latency flags and returns the stdout stream
func (p *FFmpegProcess) Start(ctx context.Context) (io.ReadCloser, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning {
		return p.stdoutPipe, nil
	}

	cmdCtx, cancel := context.WithCancel(ctx)
	p.cancelFunc = cancel

	// Low-latency FFmpeg parameters for RTSP over TCP -> MPEG1 video stream over MPEG-TS
	args := []string{
		"-rtsp_transport", "tcp",         // Force TCP for reliable packet delivery
		"-stimeout", "5000000",             // 5 seconds network timeout (in microseconds)
		"-i", p.rtspURL,                    // Input RTSP stream URL
		"-f", "mpegts",                     // Container format: MPEG-TS
		"-codec:v", "mpeg1video",           // Video codec: MPEG1 (supported natively by JSMpeg in canvas)
		"-s", "960x540",                    // Default scale resolution for smooth multi-grid rendering
		"-b:v", "1000k",                    // Video bitrate: 1000 kbps (balanced quality and low latency)
		"-bf", "0",                         // No B-frames to minimize latency
		"-r", "25",                         // Frame rate: 25 fps
		"-an",                              // Disable audio to minimize bandwidth and processing overhead
		"-tune", "zerolatency",             // Zero-latency tuning
		"-preset", "ultrafast",             // Ultrafast encoding preset
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

	// Capture stderr for debugging and failure diagnosis
	p.cmd.Stderr = p.stderrBuf

	if err := p.cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	p.isRunning = true
	log.Printf("[FFmpeg] Started worker for RTSP: %s (PID: %d)", p.rtspURL, p.cmd.Process.Pid)
	return stdout, nil
}

// Stop gracefully terminates the FFmpeg process
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
		// Attempt clean exit, then force kill if needed
		done := make(chan error, 1)
		go func() {
			done <- p.cmd.Wait()
		}()

		select {
		case <-done:
		case <-time.After(1 * time.Second):
			_ = p.cmd.Process.Kill()
		}
		log.Printf("[FFmpeg] Stopped worker for: %s", p.rtspURL)
	}

	return nil
}

// LastStderr returns the most recent error output from FFmpeg
func (p *FFmpegProcess) LastStderr() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	raw := p.stderrBuf.String()
	lines := strings.Split(raw, "\n")
	if len(lines) > 5 {
		return strings.Join(lines[len(lines)-5:], " ")
	}
	return raw
}

// IsRunning reports whether the process is actively executing
func (p *FFmpegProcess) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isRunning
}
