package model

// SendSMSRequest represents an SMS sending request
type SendSMSRequest struct {
	To        string `json:"to" validate:"required"`
	Message   string `json:"message" validate:"required"`
	Port      string `json:"port,omitempty"`
	BaudRate  int    `json:"baud_rate,omitempty"`
	Timeout   int    `json:"timeout,omitempty"`
	Mode      string `json:"mode,omitempty"` // "text" or "pdu", default "text"
	Priority  string `json:"priority,omitempty"` // "normal", "high", "urgent"
}

// SendSMSResponse represents the response of SMS sending
type SendSMSResponse struct {
	Success   bool     `json:"success"`
	MessageID string   `json:"message_id,omitempty"`
	Steps     []string `json:"steps,omitempty"`
	Error     string   `json:"error,omitempty"`
	Duration  string   `json:"duration"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Uptime    string `json:"uptime,omitempty"`
}

// PortStatusResponse represents port status response
type PortStatusResponse struct {
	Port      string `json:"port"`
	Available bool   `json:"available"`
	InUse     bool   `json:"in_use"`
	Error     string `json:"error,omitempty"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Code      int    `json:"code"`
	Timestamp string `json:"timestamp"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp string      `json:"timestamp"`
}
