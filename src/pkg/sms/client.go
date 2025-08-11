package sms

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
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
	log.Printf("SMS Client: Starting SendViaPDU - Port: %s, BaudRate: %d, To: %s", portName, baudRate, to)
	var steps []string

	// Check if port exists
	if _, err := os.Stat(portName); os.IsNotExist(err) {
		log.Printf("SMS Client: Port %s does not exist", portName)
		return steps, "", fmt.Errorf("port %s does not exist", portName)
	}

	steps = append(steps, fmt.Sprintf("Opening port %s at %d baud", portName, baudRate))

	mode := &serial.Mode{
		BaudRate: baudRate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		log.Printf("SMS Client: Failed to open port %s: %v", portName, err)
		return steps, "", fmt.Errorf("failed to open port: %w", err)
	}
	defer port.Close()

	steps = append(steps, "Port opened successfully")
	log.Printf("SMS Client: Port opened successfully")

	// Initialize modem
	log.Printf("SMS Client: Initializing modem...")
	if err := c.initializeModem(ctx, port, &steps); err != nil {
		log.Printf("SMS Client: Modem initialization failed: %v", err)
		return steps, "", err
	}

	// Set SMS mode to PDU
	steps = append(steps, "Setting SMS mode to PDU")
	log.Printf("SMS Client: Setting SMS mode to PDU...")
	if err := c.sendATCommand(ctx, port, "AT+CMGF=0", "OK"); err != nil {
		log.Printf("SMS Client: Failed to set PDU mode: %v", err)
		return steps, "", fmt.Errorf("failed to set PDU mode: %w", err)
	}

	// Generate PDU
	log.Printf("SMS Client: Generating PDU...")
	pdu, err := c.generatePDU(to, message)
	if err != nil {
		log.Printf("SMS Client: Failed to generate PDU: %v", err)
		return steps, "", fmt.Errorf("failed to generate PDU: %w", err)
	}

	steps = append(steps, fmt.Sprintf("Generated PDU: %s", pdu))
	log.Printf("SMS Client: PDU generated: %s", pdu)

	// Calculate PDU length
	pduLength := (len(pdu) - 2) / 2 // Subtract 2 for SMSC length and divide by 2 for hex

	// Send SMS
	steps = append(steps, fmt.Sprintf("Sending SMS command with length %d", pduLength))
	command := fmt.Sprintf("AT+CMGS=%d", pduLength)
	log.Printf("SMS Client: Sending command: %s", command)

	if err := c.sendATCommand(ctx, port, command, ">"); err != nil {
		log.Printf("SMS Client: Failed to initiate SMS send: %v", err)
		return steps, "", fmt.Errorf("failed to initiate SMS send: %w", err)
	}

	steps = append(steps, "Sending PDU data")
	log.Printf("SMS Client: Sending PDU data...")
	if err := c.sendATCommand(ctx, port, pdu+"\x1A", "OK"); err != nil {
		log.Printf("SMS Client: Failed to send SMS: %v", err)
		return steps, "", fmt.Errorf("failed to send SMS: %w", err)
	}

	steps = append(steps, "SMS sent successfully")
	messageID := fmt.Sprintf("SMS_%d", time.Now().Unix())
	log.Printf("SMS Client: SMS sent successfully - MessageID: %s", messageID)

	return steps, messageID, nil
}

// SendViaText sends SMS using text mode (easier than PDU mode)
func (c *Client) SendViaText(ctx context.Context, portName string, baudRate int, to, message string) ([]string, string, error) {
	log.Printf("SMS Client: Starting SendViaText - Port: %s, BaudRate: %d, To: %s", portName, baudRate, to)
	var steps []string

	// Check if port exists
	if _, err := os.Stat(portName); os.IsNotExist(err) {
		log.Printf("SMS Client: Port %s does not exist", portName)
		return steps, "", fmt.Errorf("port %s does not exist", portName)
	}

	steps = append(steps, fmt.Sprintf("Opening port %s at %d baud", portName, baudRate))

	mode := &serial.Mode{
		BaudRate: baudRate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		log.Printf("SMS Client: Failed to open port %s: %v", portName, err)
		return steps, "", fmt.Errorf("failed to open port: %w", err)
	}
	defer port.Close()

	steps = append(steps, "Port opened successfully")
	log.Printf("SMS Client: Port opened successfully")

	// Initialize modem
	log.Printf("SMS Client: Initializing modem...")
	if err := c.initializeModem(ctx, port, &steps); err != nil {
		log.Printf("SMS Client: Modem initialization failed: %v", err)
		return steps, "", err
	}

	// Set SMS mode to text
	steps = append(steps, "Setting SMS mode to text")
	log.Printf("SMS Client: Setting SMS mode to text...")
	if err := c.sendATCommand(ctx, port, "AT+CMGF=1", "OK"); err != nil {
		log.Printf("SMS Client: Failed to set text mode: %v", err)
		return steps, "", fmt.Errorf("failed to set text mode: %w", err)
	}

	// Format phone number with international prefix if needed
	formattedTo := to
	if !strings.HasPrefix(to, "+") {
		formattedTo = "+84" + strings.TrimPrefix(to, "0")
		log.Printf("SMS Client: Formatted phone number: %s -> %s", to, formattedTo)
	}

	// Send SMS
	steps = append(steps, fmt.Sprintf("Sending SMS to %s", formattedTo))
	command := fmt.Sprintf("AT+CMGS=\"%s\"", formattedTo)
	log.Printf("SMS Client: Sending command: %s", command)

	if err := c.sendATCommand(ctx, port, command, ">"); err != nil {
		log.Printf("SMS Client: Failed to initiate SMS send: %v", err)
		return steps, "", fmt.Errorf("failed to initiate SMS send: %w", err)
	}

	steps = append(steps, "Sending message text")
	log.Printf("SMS Client: Sending message text...")
	if err := c.sendATCommand(ctx, port, message+"\x1A", "OK"); err != nil {
		log.Printf("SMS Client: Failed to send SMS: %v", err)
		return steps, "", fmt.Errorf("failed to send SMS: %w", err)
	}

	steps = append(steps, "SMS sent successfully")
	messageID := fmt.Sprintf("SMS_%d", time.Now().Unix())
	log.Printf("SMS Client: SMS sent successfully - MessageID: %s", messageID)

	return steps, messageID, nil
}

// initializeModem initializes the modem
func (c *Client) initializeModem(ctx context.Context, port serial.Port, steps *[]string) error {
	// Test AT command
	*steps = append(*steps, "Testing modem with AT command")
	log.Printf("SMS Client: Testing modem with AT command...")
	if err := c.sendATCommand(ctx, port, "AT", "OK"); err != nil {
		log.Printf("SMS Client: Modem not responding to AT: %v", err)
		return fmt.Errorf("modem not responding: %w", err)
	}

	// Disable echo
	*steps = append(*steps, "Disabling echo")
	log.Printf("SMS Client: Disabling echo...")
	if err := c.sendATCommand(ctx, port, "ATE0", "OK"); err != nil {
		log.Printf("SMS Client: Failed to disable echo: %v", err)
		return fmt.Errorf("failed to disable echo: %w", err)
	}

	// Check network registration
	*steps = append(*steps, "Checking network registration")
	log.Printf("SMS Client: Checking network registration...")
	if err := c.sendATCommand(ctx, port, "AT+CREG?", "OK"); err != nil {
		log.Printf("SMS Client: Failed to check network: %v", err)
		return fmt.Errorf("failed to check network: %w", err)
	}

	return nil
}

// sendATCommand sends AT command and waits for expected response
func (c *Client) sendATCommand(ctx context.Context, port serial.Port, command, expected string) error {
	log.Printf("SMS Client: Sending AT command: %s, expecting: %s", command, expected)

	// Send command
	_, err := port.Write([]byte(command + "\r\n"))
	if err != nil {
		log.Printf("SMS Client: Failed to write command: %v", err)
		return err
	}

	// Wait for response - reduce timeout to 5 seconds
	timeout := time.Duration(5) * time.Second
	log.Printf("SMS Client: Waiting for response with timeout: %v", timeout)
	response, err := c.readUntil(ctx, port, timeout, expected)
	if err != nil {
		log.Printf("SMS Client: Failed to read response: %v", err)
		return err
	}

	log.Printf("SMS Client: Received response: %s", response)
	if !strings.Contains(strings.ToUpper(response), strings.ToUpper(expected)) {
		log.Printf("SMS Client: Unexpected response: %s", response)
		return fmt.Errorf("unexpected response: %s", response)
	}

	log.Printf("SMS Client: Command successful")
	return nil
}

// readUntil reads from port until expected string or timeout
func (c *Client) readUntil(ctx context.Context, port serial.Port, timeout time.Duration, expected string) (string, error) {
	deadline := time.Now().Add(timeout)
	var response strings.Builder
	reader := bufio.NewReader(port)

	log.Printf("SMS Client: Starting readUntil, timeout: %v, expected: %s", timeout, expected)

	for {
		select {
		case <-ctx.Done():
			log.Printf("SMS Client: Context cancelled: %v", ctx.Err())
			return response.String(), ctx.Err()
		default:
		}

		if time.Now().After(deadline) {
			log.Printf("SMS Client: Timeout waiting for: %s", expected)
			return response.String(), fmt.Errorf("timeout waiting for: %s", expected)
		}

		port.SetReadTimeout(100 * time.Millisecond)
		b, err := reader.ReadByte()
		if err != nil {
			// This is expected when no data is available
			continue
		}

		response.WriteByte(b)
		text := response.String()

		if strings.Contains(strings.ToUpper(text), strings.ToUpper(expected)) {
			log.Printf("SMS Client: Found expected response: %s", expected)
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
