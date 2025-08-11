package types

// SMSRequest represents the request structure for sending SMS
type SMSRequest struct {
	Type     string `json:"type"`
	Length   int    `json:"length"`
	Message  string `json:"message"`
	Phone    string `json:"phone"`
	Port     string `json:"port"`
	BaudRate int    `json:"baud_rate"`
}

// SMSResponse represents the response structure
type SMSResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	OTP     string `json:"otp,omitempty"`
}

// OTPRequest represents OTP generation request
type OTPRequest struct {
	Phone    string `json:"phone"`
	Port     string `json:"port"`
	BaudRate int    `json:"baud_rate"`
	Length   int    `json:"length"`
}

// OTPResponse represents OTP generation response
type OTPResponse struct {
	Success bool   `json:"success"`
	OTP     string `json:"otp,omitempty"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
