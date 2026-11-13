package models

import (
	"net/url"
)

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
