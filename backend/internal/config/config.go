package config

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all server configuration variables
type Config struct {
	Port               string
	Env                string
	AllowedOrigins     []string
	DefaultRTSPBaseURL string
	FFmpegPath         string
}

// Load reads configuration from environment variables with intelligent local/cloud path resolution
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] No .env file found, reading from system environment")
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		return nil, fmt.Errorf("required environment variable PORT is not set")
	}

	env := strings.TrimSpace(os.Getenv("ENV"))
	if env == "" {
		return nil, fmt.Errorf("required environment variable ENV is not set")
	}

	baseURL := strings.TrimSpace(os.Getenv("DEFAULT_RTSP_BASE_URL"))
	if baseURL == "" {
		return nil, fmt.Errorf("required environment variable DEFAULT_RTSP_BASE_URL is not set")
	}

	originsRaw := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
	if originsRaw == "" {
		return nil, fmt.Errorf("required environment variable ALLOWED_ORIGINS is not set")
	}

	var allowedOrigins []string
	for _, o := range strings.Split(originsRaw, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			allowedOrigins = append(allowedOrigins, trimmed)
		}
	}

	if len(allowedOrigins) == 0 {
		return nil, fmt.Errorf("ALLOWED_ORIGINS must contain at least one valid origin")
	}

	// Resolve FFmpeg path for both local (custom Windows path) and cloud (system PATH in Docker/Linux)
	ffmpegPath := resolveFFmpegPath()

	return &Config{
		Port:               port,
		Env:                env,
		AllowedOrigins:     allowedOrigins,
		DefaultRTSPBaseURL: strings.TrimRight(baseURL, "/"),
		FFmpegPath:         ffmpegPath,
	}, nil
}

// resolveFFmpegPath discovers the valid FFmpeg executable location
// 1. Checks explicit FFMPEG_PATH env var if provided (useful for local Windows testing)
// 2. Looks up "ffmpeg" in system PATH (standard in Docker containers & Linux cloud deployments)
func resolveFFmpegPath() string {
	if explicitPath := strings.TrimSpace(os.Getenv("FFMPEG_PATH")); explicitPath != "" {
		if path, err := exec.LookPath(explicitPath); err == nil {
			log.Printf("[Config] Using explicit FFmpeg path: %s", path)
			return path
		}
		log.Printf("[Config] Explicit FFMPEG_PATH '%s' not found, falling back to system search", explicitPath)
	}

	if systemPath, err := exec.LookPath("ffmpeg"); err == nil {
		log.Printf("[Config] Found FFmpeg in system PATH: %s", systemPath)
		return systemPath
	}

	// Default fallback to "ffmpeg" for containers where PATH is resolved at execution time
	log.Println("[Config] WARNING: 'ffmpeg' not immediately found in PATH at startup. Ensure FFmpeg is installed.")
	return "ffmpeg"
}



