package service

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"myproject/internal/utils"

	serial "go.bug.st/serial"
)

var (
	portMutex = &sync.Mutex{} // Đảm bảo không có xung đột khi sử dụng cổng
)

// SMSService handles SMS operations
type SMSService struct {
	smsSender *SMSSender
}

// NewSMSService creates a new SMS service instance
func NewSMSService() *SMSService {
	return &SMSService{
		smsSender: NewSMSSender(),
	}
}

// CheckPortAvailability checks if a serial port is available
func (s *SMSService) CheckPortAvailability(portName string) (bool, error) {
	serialManager := NewSerialManager()
	return serialManager.CheckPortAvailability(portName)
}

// SendSMSViaAT sends SMS using AT commands
func (s *SMSService) SendSMSViaAT(ctx context.Context, portName string, baudRate int, to string, message string) ([]string, string, error) {
	return s.smsSender.SendSMS(ctx, portName, baudRate, to, message)
}

// readUntilWithContext reads from serial port until expected response or timeout
func (s *SMSService) readUntilWithContext(ctx context.Context, port serial.Port, timeout time.Duration, expect ...string) (string, error) {
	deadline := time.Now().Add(timeout)
	var sb strings.Builder
	reader := bufio.NewReader(port)

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return sb.String(), ctx.Err()
		default:
		}

		// Check timeout
		if time.Now().After(deadline) {
			text := sb.String()
			if len(expect) == 0 {
				return text, nil
			}
			return text, fmt.Errorf("timeout reached after %v", timeout)
		}

		// Set short read timeout to make it non-blocking
		port.SetReadTimeout(100 * time.Millisecond)
		b, err := reader.ReadByte()
		if err != nil {
			if utils.IsTimeoutError(err) {
				continue // Try again
			}
			// For other errors, continue trying
			continue
		}

		sb.WriteByte(b)
		text := sb.String()

		// Check for expected responses
		for _, token := range expect {
			if token != "" && strings.Contains(strings.ToUpper(text), strings.ToUpper(token)) {
				return text, nil
			}
		}

		// Default behavior for AT commands
		if len(expect) == 0 {
			upperText := strings.ToUpper(text)
			if strings.Contains(upperText, "OK") || strings.Contains(upperText, "ERROR") {
				return text, nil
			}
		}
	}
}
