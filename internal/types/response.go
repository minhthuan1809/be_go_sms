package types

// Response body with detailed information
type SendSMSResponse struct {
	Steps     []string `json:"steps"`
	Success   bool     `json:"success"`
	Error     string   `json:"error,omitempty"`
	MessageID string   `json:"message_id,omitempty"`
	Duration  string   `json:"duration,omitempty"`
}

// Health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// Port status response
type PortStatusResponse struct {
	Port      string `json:"port"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}
