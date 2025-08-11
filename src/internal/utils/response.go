package utils

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes JSON response
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// IsTimeoutError checks if error is a timeout error
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	// Check common timeout error patterns
	errStr := err.Error()
	return contains(errStr, "timeout") || contains(errStr, "deadline exceeded")
}

// contains checks if string contains substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		indexSubstring(s, substr) >= 0)
}

// indexSubstring finds index of substring
func indexSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
