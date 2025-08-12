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
	Balance   string `json:"balance,omitempty"`
	Error     string `json:"error,omitempty"`
}

// PortInfo represents detailed port information
type PortInfo struct {
	Port        string   `json:"port"`
	DeviceName  string   `json:"device_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Msisdn      string   `json:"msisdn,omitempty"`
	Balance     string   `json:"balance,omitempty"`
	Packages    []string `json:"packages,omitempty"`
	Available   bool     `json:"available"`
	InUse       bool     `json:"in_use"`
	Error       string   `json:"error,omitempty"`
}

// SendSMSResponse represents the response from sending an SMS
type SendSMSResponse struct {
	Success   bool     `json:"success"`
	MessageID string   `json:"message_id,omitempty"`
	Error     string   `json:"error,omitempty"`
	Steps     []string `json:"steps,omitempty"`
	Duration  string   `json:"duration,omitempty"`
	Mode      string   `json:"mode,omitempty"` // "text" or "pdu"
	To        string   `json:"to,omitempty"`
	Message   string   `json:"message,omitempty"`
	Timestamp string   `json:"timestamp,omitempty"`
}

// DeviceInfo represents detailed device information including SIM details
type DeviceInfo struct {
	Port         string `json:"port"`
	BaudRate     int    `json:"baud_rate"`
	PhoneNumber  string `json:"phone_number,omitempty"`
	Balance      string `json:"balance,omitempty"`
	NetworkType  string `json:"network_type,omitempty"`
	Operator     string `json:"operator,omitempty"`
	SignalLevel  int    `json:"signal_level,omitempty"`
	IMEI         string `json:"imei,omitempty"`
	IMSI         string `json:"imsi,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
	Version      string `json:"version,omitempty"`
	Connected    bool   `json:"connected"`
	Error        string `json:"error,omitempty"`
	Timestamp    string `json:"timestamp"`
}
