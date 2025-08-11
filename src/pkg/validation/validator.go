package validation

import (
	"fmt"
	"regexp"
	"strings"

	"sms-gateway/src/internal/model"
)

// ValidatePhoneNumber validates phone number format
func ValidatePhoneNumber(phone string) error {
	// Remove any non-digit characters except +
	cleanPhone := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	// Check if it's a valid international format
	if strings.HasPrefix(cleanPhone, "+") {
		if len(cleanPhone) < 10 || len(cleanPhone) > 16 {
			return fmt.Errorf("invalid phone number length")
		}
	} else {
		if len(cleanPhone) < 9 || len(cleanPhone) > 15 {
			return fmt.Errorf("invalid phone number length")
		}
	}

	// Check if it contains only digits (and + at the beginning)
	matched, _ := regexp.MatchString(`^\+?\d+$`, cleanPhone)
	if !matched {
		return fmt.Errorf("phone number contains invalid characters")
	}

	return nil
}

// ValidateSMSMessage validates SMS message content
func ValidateSMSMessage(message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("message cannot be empty")
	}

	if len(message) > 160 {
		return fmt.Errorf("message too long (max 160 characters)")
	}

	return nil
}

// ValidateSendSMSRequest validates the entire SMS request
func ValidateSendSMSRequest(req *model.SendSMSRequest) error {
	// Validate phone number
	if err := ValidatePhoneNumber(req.To); err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Validate message
	if err := ValidateSMSMessage(req.Message); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	// Validate baud rate if provided
	if req.BaudRate != 0 {
		validBaudRates := []int{9600, 19200, 38400, 57600, 115200}
		valid := false
		for _, rate := range validBaudRates {
			if req.BaudRate == rate {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid baud rate: %d", req.BaudRate)
		}
	}

	// Validate timeout if provided
	if req.Timeout != 0 && (req.Timeout < 5 || req.Timeout > 300) {
		return fmt.Errorf("timeout must be between 5 and 300 seconds")
	}

	return nil
}
