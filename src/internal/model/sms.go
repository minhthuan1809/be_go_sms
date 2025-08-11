package model

import (
	"time"
)

// SMS represents an SMS message
type SMS struct {
	ID          string     `json:"id"`
	From        string     `json:"from"`
	To          string     `json:"to"`
	Message     string     `json:"message"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	ErrorMsg    string     `json:"error_msg,omitempty"`
}

// SMSStatus constants
const (
	StatusPending   = "pending"
	StatusSending   = "sending"
	StatusSent      = "sent"
	StatusDelivered = "delivered"
	StatusFailed    = "failed"
)

// OTP represents an OTP entry
type OTP struct {
	ID        string    `json:"id"`
	Phone     string    `json:"phone"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// ModemInfo represents modem information
type ModemInfo struct {
	Port         string `json:"port"`
	BaudRate     int    `json:"baud_rate"`
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
	Version      string `json:"version,omitempty"`
	IMEI         string `json:"imei,omitempty"`
	Signal       int    `json:"signal,omitempty"`
	Connected    bool   `json:"connected"`
}

// PortStatus represents port availability status
type PortStatus struct {
	Port      string `json:"port"`
	Available bool   `json:"available"`
	InUse     bool   `json:"in_use"`
	Error     string `json:"error,omitempty"`
}
