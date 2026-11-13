package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all server configuration variables
type Config struct {
	Port               string
	Env                string
	AllowedOrigins     []string
	FFmpegPath         string
	DefaultRTSPBaseURL string
}

// Load reads configuration strictly from environment variables without hardcoded fallback defaults
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

	ffmpegPath := strings.TrimSpace(os.Getenv("FFMPEG_PATH"))
	if ffmpegPath == "" {
		return nil, fmt.Errorf("required environment variable FFMPEG_PATH is not set")
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

	return &Config{
		Port:               port,
		Env:                env,
		AllowedOrigins:     allowedOrigins,
		FFmpegPath:         ffmpegPath,
		DefaultRTSPBaseURL: strings.TrimRight(baseURL, "/"),
	}, nil
}

