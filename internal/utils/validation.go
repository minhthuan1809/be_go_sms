package utils

import (
	"errors"
	"regexp"
	"strings"

	"myproject/internal/types"
)

// ValidateRequest validates and sets defaults for SMS request
func ValidateRequest(req *types.SendSMSRequest) error {
	// Set defaults
	if strings.TrimSpace(req.Port) == "" {
		req.Port = "/dev/ttyUSB0" // Default cho Linux, Windows thường là COM3, COM4...
	}
	if req.BaudRate == 0 {
		req.BaudRate = 115200
	}
	if req.Timeout == 0 {
		req.Timeout = 30
	}

	// Validate required fields
	if strings.TrimSpace(req.To) == "" {
		return errors.New("field 'to' is required")
	}
	if strings.TrimSpace(req.Message) == "" {
		return errors.New("field 'message' is required")
	}

	// Validate phone number format
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	if !phoneRegex.MatchString(strings.ReplaceAll(req.To, " ", "")) {
		return errors.New("invalid phone number format")
	}

	// Validate message length (SMS limit is usually 160 characters for GSM 7-bit)
	if len(req.Message) > 160 {
		return errors.New("message too long (max 160 characters)")
	}

	// Validate baud rate
	validBaudRates := []int{9600, 19200, 38400, 57600, 115200, 230400}
	valid := false
	for _, rate := range validBaudRates {
		if req.BaudRate == rate {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("invalid baud rate")
	}

	return nil
}
