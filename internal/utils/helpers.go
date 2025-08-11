package utils

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// WriteJSON writes JSON response to HTTP writer
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// Sanitize cleans up response for logging
func Sanitize(s string) string {
	// Clean up response for logging
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	if len(s) > 150 {
		return s[:150] + "..."
	}
	return s
}

// IsTimeoutError checks if error is a timeout error
func IsTimeoutError(err error) bool {
	// This is a simple check - in real implementation you might want to check specific error types
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "i/o timeout") ||
		errors.Is(err, io.EOF)
}
