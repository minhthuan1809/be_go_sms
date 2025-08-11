package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateID generates a unique ID
func GenerateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateMessageID generates a message ID with timestamp
func GenerateMessageID() string {
	timestamp := time.Now().Unix()
	id := GenerateID()
	return fmt.Sprintf("SMS_%d_%s", timestamp, id[:8])
}

// FormatDuration formats duration to human readable string
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fμs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1000000)
	}
	return d.String()
}

// SanitizeString removes unwanted characters from string
func SanitizeString(s string) string {
	// Remove control characters and normalize whitespace
	result := ""
	for _, r := range s {
		if r >= 32 && r < 127 { // Printable ASCII
			result += string(r)
		}
	}
	return result
}
