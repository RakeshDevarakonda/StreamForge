package config

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all server configuration variables
type Config struct {
	Port           string
	Env            string
	AllowedOrigins []string
	FFmpegPath     string
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
		Port:           port,
		Env:            env,
		AllowedOrigins: allowedOrigins,
		FFmpegPath:     ffmpegPath,
	}, nil
}

// resolveFFmpegPath discovers the valid FFmpeg executable location
// 1. Checks explicit FFMPEG_PATH env var if provided
// 2. Looks up "ffmpeg" in current process PATH
// 3. Fallback: Automatically scans Windows WinGet package directory
// 4. Default fallback to "ffmpeg" (for Linux/Docker cloud environments)
func resolveFFmpegPath() string {
	if explicitPath := strings.TrimSpace(os.Getenv("FFMPEG_PATH")); explicitPath != "" {
		if path, err := exec.LookPath(explicitPath); err == nil {
			log.Printf("[Config] Using explicit FFmpeg path: %s", path)
			return path
		}
		if _, err := os.Stat(explicitPath); err == nil {
			log.Printf("[Config] Found FFmpeg at explicit file path: %s", explicitPath)
			return explicitPath
		}
	}

	if systemPath, err := exec.LookPath("ffmpeg"); err == nil {
		log.Printf("[Config] Found FFmpeg in system PATH: %s", systemPath)
		return systemPath
	}

	// Windows fallback: Check WinGet default installation folder if terminal hasn't refreshed PATH
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		winGetPath := filepath.Join(localAppData, "Microsoft", "WinGet", "Packages")
		var discoveredPath string
		_ = filepath.Walk(winGetPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.EqualFold(info.Name(), "ffmpeg.exe") {
				discoveredPath = path
				return filepath.SkipAll
			}
			return nil
		})

		if discoveredPath != "" {
			log.Printf("[Config] Automatically discovered Windows WinGet FFmpeg: %s", discoveredPath)
			return discoveredPath
		}
	}

	log.Println("[Config] Using default system command: ffmpeg")
	return "ffmpeg"
}



