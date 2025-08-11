package model

// Request for sending message
type SentMessageRequest struct {
    Type      string `json:"type"`       // OTP || TEXT
    Length    int    `json:"length"`     // OTP length (required for OTP type)
    Message   string `json:"message"`    // Message content
    Phone     string `json:"phone"`      // Phone number to send to
    Port      string `json:"port"`       // Serial port (e.g., /dev/ttyUSB0)
    BaudRate  int    `json:"baud_rate"`  // Serial port baud rate
}

// Response for sent message
type SentMessageResponse struct {
    Success bool   `json:"success"`
    Error   string `json:"error,omitempty"`
    OTP     string `json:"otp,omitempty"`
    Message string `json:"message,omitempty"`
}

// Internal SMS request for modem communication
type SendSMSRequest struct {
    Port     string `json:"port"`
    BaudRate int    `json:"baud_rate"`
    To       string `json:"to"`
    Message  string `json:"message"`
    Timeout  int    `json:"timeout"`
}
