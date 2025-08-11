package sms

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"sms-gateway/src/internal/config"

	serial "go.bug.st/serial"
)

// Client handles SMS operations
type Client struct {
	config *config.Config
}

// NewClient creates a new SMS client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
	}
}

// SendViaPDU sends SMS using PDU mode
func (c *Client) SendViaPDU(ctx context.Context, portName string, baudRate int, to, message string) ([]string, string, error) {
	var steps []string

	steps = append(steps, fmt.Sprintf("Opening port %s at %d baud", portName, baudRate))

	mode := &serial.Mode{
		BaudRate: baudRate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return steps, "", fmt.Errorf("failed to open port: %w", err)
	}
	defer port.Close()

	steps = append(steps, "Port opened successfully")

	// Initialize modem
	if err := c.initializeModem(ctx, port, &steps); err != nil {
		return steps, "", err
	}

	// Set SMS mode to PDU
	steps = append(steps, "Setting SMS mode to PDU")
	if err := c.sendATCommand(ctx, port, "AT+CMGF=0", "OK"); err != nil {
		return steps, "", fmt.Errorf("failed to set PDU mode: %w", err)
	}

	// Generate PDU
	pdu, err := c.generatePDU(to, message)
	if err != nil {
		return steps, "", fmt.Errorf("failed to generate PDU: %w", err)
	}

	steps = append(steps, fmt.Sprintf("Generated PDU: %s", pdu))

	// Calculate PDU length
	pduLength := (len(pdu) - 2) / 2 // Subtract 2 for SMSC length and divide by 2 for hex

	// Send SMS
	steps = append(steps, fmt.Sprintf("Sending SMS command with length %d", pduLength))
	command := fmt.Sprintf("AT+CMGS=%d", pduLength)

	if err := c.sendATCommand(ctx, port, command, ">"); err != nil {
		return steps, "", fmt.Errorf("failed to initiate SMS send: %w", err)
	}

	steps = append(steps, "Sending PDU data")
	if err := c.sendATCommand(ctx, port, pdu+"\x1A", "OK"); err != nil {
		return steps, "", fmt.Errorf("failed to send SMS: %w", err)
	}

	steps = append(steps, "SMS sent successfully")
	messageID := fmt.Sprintf("SMS_%d", time.Now().Unix())

	return steps, messageID, nil
}

// initializeModem initializes the modem
func (c *Client) initializeModem(ctx context.Context, port serial.Port, steps *[]string) error {
	// Test AT command
	*steps = append(*steps, "Testing modem with AT command")
	if err := c.sendATCommand(ctx, port, "AT", "OK"); err != nil {
		return fmt.Errorf("modem not responding: %w", err)
	}

	// Disable echo
	*steps = append(*steps, "Disabling echo")
	if err := c.sendATCommand(ctx, port, "ATE0", "OK"); err != nil {
		return fmt.Errorf("failed to disable echo: %w", err)
	}

	// Check network registration
	*steps = append(*steps, "Checking network registration")
	if err := c.sendATCommand(ctx, port, "AT+CREG?", "OK"); err != nil {
		return fmt.Errorf("failed to check network: %w", err)
	}

	return nil
}

// sendATCommand sends AT command and waits for expected response
func (c *Client) sendATCommand(ctx context.Context, port serial.Port, command, expected string) error {
	// Send command
	_, err := port.Write([]byte(command + "\r\n"))
	if err != nil {
		return err
	}

	// Wait for response
	timeout := time.Duration(10) * time.Second
	response, err := c.readUntil(ctx, port, timeout, expected)
	if err != nil {
		return err
	}

	if !strings.Contains(strings.ToUpper(response), strings.ToUpper(expected)) {
		return fmt.Errorf("unexpected response: %s", response)
	}

	return nil
}

// readUntil reads from port until expected string or timeout
func (c *Client) readUntil(ctx context.Context, port serial.Port, timeout time.Duration, expected string) (string, error) {
	deadline := time.Now().Add(timeout)
	var response strings.Builder
	reader := bufio.NewReader(port)

	for {
		select {
		case <-ctx.Done():
			return response.String(), ctx.Err()
		default:
		}

		if time.Now().After(deadline) {
			return response.String(), fmt.Errorf("timeout waiting for: %s", expected)
		}

		port.SetReadTimeout(100 * time.Millisecond)
		b, err := reader.ReadByte()
		if err != nil {
			continue
		}

		response.WriteByte(b)
		text := response.String()

		if strings.Contains(strings.ToUpper(text), strings.ToUpper(expected)) {
			return text, nil
		}
	}
}

// generatePDU generates PDU for SMS
func (c *Client) generatePDU(to, message string) (string, error) {
	// Simple PDU generation (this is a basic implementation)
	// For production, you might want to use a more robust PDU library

	// SMSC length (empty)
	pdu := "00"

	// SMS-SUBMIT type
	pdu += "11"

	// Message reference
	pdu += "00"

	// Destination address length
	pdu += fmt.Sprintf("%02X", len(to))

	// Destination address type (international)
	pdu += "91"

	// Destination address (reverse pairs and pad)
	paddedTo := to
	if len(paddedTo)%2 == 1 {
		paddedTo += "F"
	}

	for i := 0; i < len(paddedTo); i += 2 {
		if i+1 < len(paddedTo) {
			pdu += string(paddedTo[i+1]) + string(paddedTo[i])
		}
	}

	// Protocol identifier
	pdu += "00"

	// Data coding scheme
	pdu += "00"

	// User data length
	pdu += fmt.Sprintf("%02X", len(message))

	// User data (convert to hex)
	for _, char := range message {
		pdu += fmt.Sprintf("%02X", int(char))
	}

	return pdu, nil
}
